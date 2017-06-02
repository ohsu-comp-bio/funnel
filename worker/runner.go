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

func NewDefaultRunner(conf config.Worker, taskID string) Runner {

	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, taskID)
	// TODO handle error
	svc, _ := newRPCTask(conf, taskID)

  return &taskRunner{
		conf:   conf,
		mapper: NewFileMapper(baseDir),
		store:  storage.Storage{},
		taskID: taskID,
		svc:    svc,
	}
}

func NewLoggerRunner(task *tes.Task, conf config.Worker) Runner {
	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, task.Id)

	return &taskRunner{
		conf:   conf,
		mapper: NewFileMapper(baseDir),
		store:  storage.Storage{},
		taskID: task.Id,
		svc:    NewLoggerTask(task),
	}
}

// taskRunner helps collect data used across many helper methods.
type taskRunner struct {
	conf   config.Worker
	mapper *FileMapper
	store  storage.Storage
	taskID string
	svc    TaskService
}

// TODO document behavior of slow consumer of task log updates
func (r *taskRunner) Run(pctx context.Context) {

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

	log := logger.Sub("runner", "workerID", r.conf.ID, "taskID", r.taskID)
	log.Debug("Run")

	task, run.syserr = r.svc.Task()

	r.svc.StartTime(time.Now())
	// Run the final logging/state steps in a deferred function
	// to ensure they always run, even if there's a missed error.
	defer func() {
		r.svc.EndTime(time.Now())

		switch {
		case run.taskCanceled:
			// The task was canceled.
			r.svc.SetState(tes.State_CANCELED)
		case run.execerr != nil:
			// One of the executors failed
			r.svc.SetState(tes.State_ERROR)
		case run.syserr != nil:
			// Something else failed
			// TODO should we do something special for run.err == context.Canceled?
			r.svc.SetState(tes.State_SYSTEM_ERROR)
		default:
			r.svc.SetState(tes.State_COMPLETE)
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
		dir, run.syserr = filepath.Abs(r.conf.WorkDir)
	}
	if run.ok() {
		run.syserr = util.EnsureDir(dir)
	}

	// Prepare file mapper, which maps task file URLs to host filesystem paths
	if run.ok() {
		run.syserr = r.mapper.MapTask(task)
	}

	// Grab the IP address of this host. Used to send task metadata updates.
	var ip string
	if run.ok() {
		ip, run.syserr = externalIP()
	}

	// Configure a task-specific storage backend.
	// This provides download/upload for inputs/outputs.
	if run.ok() {
		r.store, run.syserr = r.store.WithConfig(r.conf.Storage)
	}

	if run.ok() {
		run.syserr = r.validateInputs()
	}

	if run.ok() {
		run.syserr = r.validateOutputs()
	}

	// Download inputs
	for _, input := range r.mapper.Inputs {
		if run.ok() {
			run.syserr = r.store.Get(ctx, input.Url, input.Path, input.Type)
		}
	}

	if run.ok() {
		r.svc.SetState(tes.State_RUNNING)
	}

	// Run steps
	for i, d := range task.Executors {
		s := &stepRunner{
			TaskID:     task.Id,
			Conf:       r.conf,
			Num:        i,
			Log:        log.WithFields("step", i),
			TaskLogger: r.svc,
			IP:         ip,
			Cmd: &DockerCmd{
				ImageName:     d.ImageName,
				Cmd:           d.Cmd,
				Environ:       d.Environ,
				Volumes:       r.mapper.Volumes,
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
	for _, output := range r.mapper.Outputs {
		if run.ok() {
			r.fixLinks(output.Path)
			var out []*tes.OutputFileLog
			out, run.syserr = r.store.Put(ctx, output.Url, output.Path, output.Type)
			outputs = append(outputs, out...)
		}
	}

	if run.ok() {
		r.svc.Outputs(outputs)
	}
}

// fixLinks walks the output paths, fixing cases where a symlink is
// broken because it's pointing to a path inside a container volume.
func (r *taskRunner) fixLinks(basepath string) {
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
				mapped, err := r.mapper.HostPath(src)
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
func (r *taskRunner) openStepLogs(s *stepRunner, d *tes.Executor) error {

	// Find the path for task stdin
	var err error
	if d.Stdin != "" {
		s.Cmd.Stdin, err = r.mapper.OpenHostFile(d.Stdin)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Cmd.Stdout, err = r.mapper.CreateHostFile(d.Stdout)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Cmd.Stderr, err = r.mapper.CreateHostFile(d.Stderr)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}
	}
	return nil
}

// Validate the input downloads
func (r *taskRunner) validateInputs() error {
	for _, input := range r.mapper.Inputs {
		if !r.store.Supports(input.Url, input.Path, input.Type) {
			return fmt.Errorf("Input download not supported by storage: %v", input)
		}
	}
	return nil
}

// Validate the output uploads
func (r *taskRunner) validateOutputs() error {
	for _, output := range r.mapper.Outputs {
		if !r.store.Supports(output.Url, output.Path, output.Type) {
			return fmt.Errorf("Output upload not supported by storage: %v", output)
		}
	}
	return nil
}

func (r *taskRunner) pollForCancel(ctx context.Context, f func()) context.Context {
	taskctx, cancel := context.WithCancel(ctx)

	// Start a goroutine that polls the server to watch for a canceled state.
	// If a cancel state is found, "taskctx" is canceled.
	go func() {
		ticker := time.NewTicker(r.conf.UpdateRate)
		defer ticker.Stop()

		for {
			select {
			case <-taskctx.Done():
				return
			case <-ticker.C:
				state := r.svc.State()
				if tes.TerminalState(state) {
					cancel()
					f()
				}
			}
		}
	}()
	return taskctx
}

// NoopTaskRunner is useful during testing for creating a worker with a TaskRunner
// that doesn't do anything.
func NoopTaskRunner(ctx context.Context, c config.Worker, taskID string) {
}
