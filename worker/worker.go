package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/version"
	"os"
	"path/filepath"
	"time"
)

// DefaultWorker is the default task worker, which follows a basic,
// sequential process of task initialization, execution, finalization,
// and logging.
type DefaultWorker struct {
	Conf       config.Worker
	Mapper     *FileMapper
	Store      storage.Storage
	TaskReader TaskReader
	Event      *events.TaskWriter
}

// Run runs the Worker.
// TODO document behavior of slow consumer of task log updates
func (r *DefaultWorker) Run(pctx context.Context) {

	// The code here is verbose, but simple; mainly loops and simple error checking.
	//
	// The steps are:
	// - prepare the working directory
	// - map the task files to the working directory
	// - log the IP address
	// - set up the storage configuration
	// - validate input and output files
	// - download inputs
	// - run the steps (docker)
	// - upload the outputs

	var run helper
	var task *tes.Task

	r.Event.Info("Version", version.LogFields()...)
	r.Event.State(tes.State_INITIALIZING)
	r.Event.StartTime(time.Now())

	if name, err := os.Hostname(); err == nil {
		r.Event.Info("Hostname", "name", name)
	}

	task, run.syserr = r.TaskReader.Task()

	// Run the final logging/state steps in a deferred function
	// to ensure they always run, even if there's a missed error.
	defer func() {
		r.Event.EndTime(time.Now())

		switch {
		case run.taskCanceled:
			// The task was canceled.
			r.Event.Info("Canceled")
			r.Event.State(tes.State_CANCELED)
		case run.execerr != nil:
			// One of the executors failed
			r.Event.Error("Exec error", "error", run.execerr)
			r.Event.State(tes.State_EXECUTOR_ERROR)
		case run.syserr != nil:
			// Something else failed
			// TODO should we do something special for run.err == context.Canceled?
			r.Event.Error("System error", "error", run.syserr)
			r.Event.State(tes.State_SYSTEM_ERROR)
		default:
			r.Event.State(tes.State_COMPLETE)
		}

		// cleanup workdir
		if !r.Conf.LeaveWorkDir {
			r.Mapper.Cleanup()
		}
	}()

	// Recover from panics
	defer handlePanic(func(e error) {
		fmt.Printf("%#v", e)
		run.syserr = e
	})

	ctx := r.pollForCancel(pctx, func() {
		run.taskCanceled = true
	})
	run.ctx = ctx

	// Prepare file mapper, which maps task file URLs to host filesystem paths
	if run.ok() {
		run.syserr = r.Mapper.MapTask(task)
	}

	if run.ok() {
		run.syserr = r.validateInputs()
	}

	if run.ok() {
		run.syserr = r.validateOutputs()
	}

	// Download inputs
	for _, input := range r.Mapper.Inputs {
		if run.ok() {
			r.Event.Info("Starting download", "url", input.Url)
			err := r.Store.Get(ctx, input.Url, input.Path, input.Type)
			if err != nil {
				run.syserr = err
				r.Event.Error("Download failed", "url", input.Url, "error", err)
			} else {
				r.Event.Info("Download finished", "url", input.Url)
			}
		}
	}

	if run.ok() {
		r.Event.State(tes.State_RUNNING)
	}

	// Run steps
	for i, d := range task.Executors {
		s := &stepWorker{
			Conf:  r.Conf,
			Event: r.Event.NewExecutorWriter(uint32(i)),
			Command: &DockerCommand{
				Image:         d.Image,
				Command:       d.Command,
				Env:           d.Env,
				Volumes:       r.Mapper.Volumes,
				Workdir:       d.Workdir,
				ContainerName: fmt.Sprintf("%s-%d", task.Id, i),
				// TODO make RemoveContainer configurable
				RemoveContainer: true,
				Event:           r.Event.NewExecutorWriter(uint32(i)),
			},
		}

		// Opens stdin/out/err files and updates those fields on "cmd".
		if run.ok() {
			run.syserr = r.openStepLogs(s, d)
		}

		if run.ok() {
			run.execerr = s.Run(ctx)
		}
	}

	// Upload outputs
	var outputs []*tes.OutputFileLog
	for _, output := range r.Mapper.Outputs {
		if run.ok() {
			r.Event.Info("Starting upload", "url", output.Url)
			r.fixLinks(output.Path)
			out, err := r.Store.Put(ctx, output.Url, output.Path, output.Type)
			if err != nil {
				run.syserr = err
				r.Event.Error("Upload failed", "url", output.Url, "error", err)
			} else {
				r.Event.Info("Upload finished", "url", output.Url)
			}
			outputs = append(outputs, out...)
		}
	}
	// unmap paths for OutputFileLog
	for _, o := range outputs {
		o.Path = r.Mapper.ContainerPath(o.Path)
	}

	if run.ok() {
		r.Event.Outputs(outputs)
	}
}

// fixLinks walks the output paths, fixing cases where a symlink is
// broken because it's pointing to a path inside a container volume.
func (r *DefaultWorker) fixLinks(basepath string) {
	filepath.Walk(basepath, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			// There's an error, so be safe and give up on this file
			return nil
		}

		// Only bother to check symlinks
		if f.Mode()&os.ModeSymlink != 0 {
			// Test if the file can be opened because it doesn't exist
			fh, rerr := os.Open(p)
			fh.Close()

			if rerr != nil && os.IsNotExist(rerr) {

				// Get symlink source path
				src, err := os.Readlink(p)
				if err != nil {
					return nil
				}
				// Map symlink source (possible container path) to host path
				mapped, err := r.Mapper.HostPath(src)
				if err != nil {
					return nil
				}

				// Check whether the mapped path exists
				fh, err := os.Open(mapped)
				fh.Close()

				// If the mapped path exists, fix the symlink
				if err == nil {
					err := os.Remove(p)
					if err != nil {
						return nil
					}
					os.Symlink(mapped, p)
				}
			}
		}
		return nil
	})
}

// openLogs opens/creates the logs files for a step and updates those fields.
func (r *DefaultWorker) openStepLogs(s *stepWorker, d *tes.Executor) error {

	// Find the path for task stdin
	var err error
	if d.Stdin != "" {
		s.Command.Stdin, err = r.Mapper.OpenHostFile(d.Stdin)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Command.Stdout, err = r.Mapper.CreateHostFile(d.Stdout)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Command.Stderr, err = r.Mapper.CreateHostFile(d.Stderr)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	return nil
}

// Validate the input downloads
func (r *DefaultWorker) validateInputs() error {
	for _, input := range r.Mapper.Inputs {
		err := r.Store.SupportsGet(input.Url, input.Type)
		if err != nil {
			return fmt.Errorf("Input download not supported by storage: %v", err)
		}
	}
	return nil
}

// Validate the output uploads
func (r *DefaultWorker) validateOutputs() error {
	for _, output := range r.Mapper.Outputs {
		err := r.Store.SupportsPut(output.Url, output.Type)
		if err != nil {
			return fmt.Errorf("Output upload not supported by storage: %v", err)
		}
	}
	return nil
}

func (r *DefaultWorker) pollForCancel(ctx context.Context, f func()) context.Context {
	taskctx, cancel := context.WithCancel(ctx)

	// Start a goroutine that polls the server to watch for a canceled state.
	// If a cancel state is found, "taskctx" is canceled.
	go func() {
		ticker := time.NewTicker(r.Conf.UpdateRate)
		defer ticker.Stop()

		for {
			select {
			case <-taskctx.Done():
				return
			case <-ticker.C:
				state, _ := r.TaskReader.State()
				if tes.TerminalState(state) {
					cancel()
					f()
				}
			}
		}
	}()
	return taskctx
}
