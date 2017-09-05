package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"path"
	"path/filepath"
	"time"
)

// NewDefaultRunner returns the default task runner used by Funnel,
// which uses gRPC to read/write task details.
func NewDefaultRunner(conf config.Worker, taskID string) Runner {
	log := logger.Sub("runner", "workerID", conf.ID, "taskID", taskID)

	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, taskID)

	// TODO handle error
	client, _ := newTaskClient(conf)

	return &DefaultRunner{
		Conf:   conf,
		Mapper: NewFileMapper(baseDir),
		Store:  storage.Storage{Log: log},
		Read:   &rpcTaskReader{client, taskID},
		Write:  &rpcTaskWriter{client, taskID, conf.UpdateTimeout, log},
		Log:    log,
	}
}

// DefaultRunner is the default task runner, which follows a basic,
// sequential process of task initialization, execution, finalization,
// and logging.
type DefaultRunner struct {
	Conf   config.Worker
	Mapper *FileMapper
	Store  storage.Storage
	Read   TaskReader
	Write  TaskWriter
	Log    logger.Logger
}

// Run runs the task runner.
// TODO document behavior of slow consumer of task log updates
func (r *DefaultRunner) Run(pctx context.Context) {

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

	task, run.syserr = r.Read.Task()

	r.Write.StartTime(time.Now())
	// Run the final logging/state steps in a deferred function
	// to ensure they always run, even if there's a missed error.
	defer func() {
		r.Write.EndTime(time.Now())

		switch {
		case run.taskCanceled:
			// The task was canceled.
			r.Log.Info("Canceled")
			r.Write.State(tes.State_CANCELED)
		case run.execerr != nil:
			// One of the executors failed
			r.Log.Error("Exec error", run.execerr)
			r.Write.State(tes.State_ERROR)
		case run.syserr != nil:
			// Something else failed
			// TODO should we do something special for run.err == context.Canceled?
			r.Log.Error("System error", run.syserr)
			r.Write.State(tes.State_SYSTEM_ERROR)
		default:
			r.Write.State(tes.State_COMPLETE)
		}
	}()

	// Recover from panics
	defer handlePanic(func(e error) {
		run.syserr = e
	})

	ctx := r.pollForCancel(pctx, func() {
		run.taskCanceled = true
	})
	run.ctx = ctx

	// Create working dir
	var dir string
	if run.ok() {
		dir, run.syserr = filepath.Abs(r.Conf.WorkDir)
	}
	if run.ok() {
		run.syserr = util.EnsureDir(dir)
	}

	// Prepare file mapper, which maps task file URLs to host filesystem paths
	if run.ok() {
		run.syserr = r.Mapper.MapTask(task)
	}

	// Grab the IP address of this host. Used to send task metadata updates.
	var ip string
	if run.ok() {
		ip, run.syserr = externalIP()
	}

	// Configure a task-specific storage backend.
	// This provides download/upload for inputs/outputs.
	if run.ok() {
		r.Store, run.syserr = r.Store.WithConfig(r.Conf.Storage)
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
			r.Log.Info("Starting download", "url", input.Url, "path", input.Path)

			run.syserr = r.Store.Get(ctx, input.Url, input.Path, input.Type)

			if run.syserr == nil {
				r.Log.Info("Finished file download", "url", input.Url, "hostPath", input.Path)
			}
		}
	}

	if run.ok() {
		r.Write.State(tes.State_RUNNING)
	}

	// Run steps
	for i, d := range task.Executors {
		s := &stepRunner{
			Conf:  r.Conf,
			Num:   i,
			Log:   r.Log.WithFields("step", i),
			Write: r.Write,
			IP:    ip,
			Cmd: &DockerCmd{
				ImageName:     d.ImageName,
				Cmd:           d.Cmd,
				Environ:       d.Environ,
				Volumes:       r.Mapper.Volumes,
				Workdir:       d.Workdir,
				Ports:         d.Ports,
				ContainerName: fmt.Sprintf("%s-%d", task.Id, i),
				// TODO make RemoveContainer configurable
				RemoveContainer: true,
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
			r.fixLinks(output.Path)
			var out []*tes.OutputFileLog

			r.Log.Info("Starting upload", "url", output.Url, "path", output.Path)

			out, run.syserr = r.Store.Put(ctx, output.Url, output.Path, output.Type)

			if run.ok() {
				r.Log.Info("Finished file upload", "url", output.Url, "path", output.Path)
			}
			outputs = append(outputs, out...)
		}
	}

	if run.ok() {
		r.Write.Outputs(outputs)
	}
}

// fixLinks walks the output paths, fixing cases where a symlink is
// broken because it's pointing to a path inside a container volume.
func (r *DefaultRunner) fixLinks(basepath string) {
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
func (r *DefaultRunner) openStepLogs(s *stepRunner, d *tes.Executor) error {

	// Find the path for task stdin
	var err error
	if d.Stdin != "" {
		s.Cmd.Stdin, err = r.Mapper.OpenHostFile(d.Stdin)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Cmd.Stdout, err = r.Mapper.CreateHostFile(d.Stdout)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Cmd.Stderr, err = r.Mapper.CreateHostFile(d.Stderr)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}
	return nil
}

// Validate the input downloads
func (r *DefaultRunner) validateInputs() error {
	for _, input := range r.Mapper.Inputs {
		if !r.Store.Supports(input.Url, input.Path, input.Type) {
			return fmt.Errorf("Input download not supported by storage: %v", input)
		}
	}
	return nil
}

// Validate the output uploads
func (r *DefaultRunner) validateOutputs() error {
	for _, output := range r.Mapper.Outputs {
		if !r.Store.Supports(output.Url, output.Path, output.Type) {
			return fmt.Errorf("Output upload not supported by storage: %v", output)
		}
	}
	return nil
}

func (r *DefaultRunner) pollForCancel(ctx context.Context, f func()) context.Context {
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
				state := r.Read.State()
				if tes.TerminalState(state) {
					cancel()
					f()
				}
			}
		}
	}()
	return taskctx
}
