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
	Conf        config.Worker
	Store       storage.Storage
	TaskReader  TaskReader
	EventWriter events.Writer
}

// Run runs the Worker.
// TODO document behavior of slow consumer of task log updates
func (r *DefaultWorker) Run(pctx context.Context, taskID string) (runerr error) {

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
	var event *events.TaskWriter
	var mapper *FileMapper
	var run helper
	var task *tes.Task

	// set up task specific utilities
	event = events.NewTaskWriter(taskID, 0, r.EventWriter)
	mapper = NewFileMapper(filepath.Join(r.Conf.WorkDir, taskID))

	event.Info("Version", version.LogFields()...)
	event.State(tes.State_INITIALIZING)
	event.StartTime(time.Now())

	if name, err := os.Hostname(); err == nil {
		event.Metadata(map[string]string{"hostname": name})
	}

	task, run.syserr = r.TaskReader.Task(pctx, taskID)

	// Run the final logging/state steps in a deferred function
	// to ensure they always run, even if there's a missed error.
	defer func() {
		event.EndTime(time.Now())

		switch {
		case run.taskCanceled:
			// The task was canceled.
			event.Info("Canceled")
			event.State(tes.State_CANCELED)
			runerr = fmt.Errorf("task canceled")
		case run.syserr != nil:
			// Something else failed
			event.Error("System error", "error", run.syserr)
			event.State(tes.State_SYSTEM_ERROR)
			runerr = run.syserr
		case run.execerr != nil:
			// One of the executors failed
			event.Error("Exec error", "error", run.execerr)
			event.State(tes.State_EXECUTOR_ERROR)
			runerr = run.execerr
		default:
			event.State(tes.State_COMPLETE)
		}

		// cleanup workdir
		if !r.Conf.LeaveWorkDir {
			mapper.Cleanup()
		}
	}()

	// Recover from panics
	defer handlePanic(func(e error) {
		fmt.Printf("%#v", e)
		run.syserr = e
	})

	ctx := r.pollForCancel(pctx, taskID, func() { run.taskCanceled = true })
	run.ctx = ctx

	// Prepare file mapper, which maps task file URLs to host filesystem paths
	if run.ok() {
		run.syserr = mapper.MapTask(task)
	}

	if run.ok() {
		run.syserr = r.validateInputs(mapper)
	}

	if run.ok() {
		run.syserr = r.validateOutputs(mapper)
	}

	// Download inputs
	for _, input := range mapper.Inputs {
		if run.ok() {
			event.Info("Download started", "url", input.Url)
			err := r.Store.Get(ctx, input.Url, input.Path, input.Type)
			if err != nil {
				if err == storage.ErrEmptyDirectory {
					event.Warn("Download finished with warning", "url", input.Url, "warning", err)
				} else {
					run.syserr = err
					event.Error("Download failed", "url", input.Url, "error", err)
				}
			} else {
				event.Info("Download finished", "url", input.Url)
			}
		}
	}

	if run.ok() {
		event.State(tes.State_RUNNING)
	}

	// Run steps
	if run.ok() {
		for i, d := range task.GetExecutors() {
			s := &stepWorker{
				Conf:  r.Conf,
				Event: event.NewExecutorWriter(uint32(i)),
				Command: &DockerCommand{
					Image:         d.Image,
					Command:       d.Command,
					Env:           d.Env,
					Volumes:       mapper.Volumes,
					Workdir:       d.Workdir,
					ContainerName: fmt.Sprintf("%s-%d", task.Id, i),
					// TODO make RemoveContainer configurable
					RemoveContainer: true,
					Event:           event.NewExecutorWriter(uint32(i)),
				},
			}

			// Opens stdin/out/err files and updates those fields on "cmd".
			if run.ok() {
				run.syserr = r.openStepLogs(mapper, s, d)
			}

			if run.ok() {
				run.execerr = s.Run(ctx)
			}
		}
	}

	// Upload outputs
	var outputs []*tes.OutputFileLog
	for _, output := range mapper.Outputs {
		if run.ok() {
			event.Info("Upload started", "url", output.Url)
			r.fixLinks(mapper, output.Path)
			out, err := r.Store.Put(ctx, output.Url, output.Path, output.Type)
			if err != nil {
				if err == storage.ErrEmptyDirectory {
					event.Warn("Upload finished with warning", "url", output.Url, "warning", err)
				} else {
					run.syserr = err
					event.Error("Upload failed", "url", output.Url, "error", err)
				}
			} else {
				event.Info("Upload finished", "url", output.Url)
			}
			outputs = append(outputs, out...)
		}
	}
	// unmap paths for OutputFileLog
	for _, o := range outputs {
		o.Path = mapper.ContainerPath(o.Path)
	}

	if run.ok() {
		event.Outputs(outputs)
	}

	return
}

// fixLinks walks the output paths, fixing cases where a symlink is
// broken because it's pointing to a path inside a container volume.
func (r *DefaultWorker) fixLinks(mapper *FileMapper, basepath string) {
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
				mapped, err := mapper.HostPath(src)
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
func (r *DefaultWorker) openStepLogs(mapper *FileMapper, s *stepWorker, d *tes.Executor) error {

	// Find the path for task stdin
	var err error
	if d.Stdin != "" {
		s.Command.Stdin, err = mapper.OpenHostFile(d.Stdin)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Command.Stdout, err = mapper.CreateHostFile(d.Stdout)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Command.Stderr, err = mapper.CreateHostFile(d.Stderr)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	return nil
}

// Validate the input downloads
func (r *DefaultWorker) validateInputs(mapper *FileMapper) error {
	for _, input := range mapper.Inputs {
		err := r.Store.SupportsGet(input.Url, input.Type)
		if err != nil {
			return fmt.Errorf("Input download not supported by storage: %v", err)
		}
	}
	return nil
}

// Validate the output uploads
func (r *DefaultWorker) validateOutputs(mapper *FileMapper) error {
	for _, output := range mapper.Outputs {
		err := r.Store.SupportsPut(output.Url, output.Type)
		if err != nil {
			return fmt.Errorf("Output upload not supported by storage: %v", err)
		}
	}
	return nil
}

func (r *DefaultWorker) pollForCancel(pctx context.Context, taskID string, cancelCallback func()) context.Context {
	taskctx, cancel := context.WithCancel(pctx)

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
				state, _ := r.TaskReader.State(taskctx, taskID)
				if tes.TerminalState(state) {
					cancel()
					cancelCallback()
				}
			}
		}
	}()
	return taskctx
}
