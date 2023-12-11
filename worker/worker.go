// Package worker contains code which executes a task.
package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/version"
)

// DefaultWorker is the default task worker, which follows a basic,
// sequential process of task initialization, execution, finalization,
// and logging.
type DefaultWorker struct {
	Executor       Executor
	Conf           config.Worker
	Store          storage.Storage
	TaskReader     TaskReader
	EventWriter    events.Writer
}

// Configuration of the task executor.
type Executor struct {
	// "docker" or "kubernetes"
	Backend   string
	// Kubernetes executor template
	Template  string
	// Kubernetes namespace
	Namespace string
}

// Run runs the Worker.
func (r *DefaultWorker) Run(pctx context.Context) (runerr error) {

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

	task, run.syserr = r.TaskReader.Task(pctx)
	// TODO if we failed to retrieve the task, we can't do anything useful.
	//      but, it's also difficult to report the failure usefully.
	if run.syserr != nil {
		r.EventWriter.WriteEvent(pctx, events.NewSystemLog("unknown", 0, 0, "error",
			"failed to get task. ID unknown", map[string]string{
				"error": run.syserr.Error(),
			}))
		runerr = run.syserr
		return
	}

	// set up task specific utilities
	event = events.NewTaskWriter(task.GetId(), 0, r.EventWriter)
	mapper = NewFileMapper(filepath.Join(r.Conf.WorkDir, task.GetId()))

	event.Info("Version", version.LogFields()...)
	event.State(tes.State_INITIALIZING)
	event.StartTime(time.Now())

	if name, err := os.Hostname(); err == nil {
		event.Metadata(map[string]string{"hostname": name})
	}

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

	ctx := r.pollForCancel(pctx, func() { run.taskCanceled = true })
	run.ctx = ctx

	// Prepare file mapper, which maps task file URLs to host filesystem paths
	if run.ok() {
		run.syserr = mapper.MapTask(task)
	}

	if run.ok() {
		run.syserr = r.validate(mapper)
	}

	// Download inputs
	if run.ok() {
		run.syserr = DownloadInputs(ctx, mapper.Inputs, r.Store, event, r.Conf.MaxParallelTransfers)
	}

	if run.ok() {
		event.State(tes.State_RUNNING)
	}

	var resources = task.GetResources()
	if resources == nil {
		resources = &tes.Resources{}
	}

	// Run steps
	if run.ok() {
		for i, d := range task.GetExecutors() {
			var command = Command{
				Image:         d.Image,
				ShellCommand:  d.Command,
				Volumes:       mapper.Volumes,
				Workdir:       d.Workdir,
				Env:           d.Env,
				Event:         event.NewExecutorWriter(uint32(i)),
			}

			var taskCommand TaskCommand
			if r.Executor.Backend == "kubernetes" {
				taskCommand = &KubernetesCommand{
					TaskId:       task.Id,
					JobId:        i, 
					StdinFile:    d.Stdin,
					TaskTemplate: r.Executor.Template,
					Namespace:    r.Executor.Namespace,
					Resources:    resources,
					Command:      command,
				}
			} else {
				taskCommand = &DockerCommand{
					ContainerName:   fmt.Sprintf("%s-%d", task.Id, i),
					//TODO Make RemoveContainer configurable
				 	RemoveContainer: true,
					Command:         command,
				}
			}

			s := &stepWorker{
				Conf:  r.Conf,
				Event: event.NewExecutorWriter(uint32(i)),
				Command: taskCommand,
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

	// Try to fix symlinks broken by docker filesystems.
	if run.ok() {
		for _, output := range mapper.Outputs {
			fixLinks(mapper, output.Path)
		}
	}

	// Upload outputs
	var outputLog []*tes.OutputFileLog
	if run.ok() {
		outputLog, run.syserr = UploadOutputs(ctx, mapper.Outputs, r.Store, event, r.Conf.MaxParallelTransfers)
	}

	// unmap paths for OutputFileLog
	for _, o := range outputLog {
		o.Path = mapper.ContainerPath(o.Path)
	}

	if len(outputLog) > 0 {
		event.Outputs(outputLog)
	}

	return
}

func (r *DefaultWorker) Close() {
	r.TaskReader.Close()
	r.EventWriter.Close()
}

// openLogs opens/creates the logs files for a step and updates those fields.
func (r *DefaultWorker) openStepLogs(mapper *FileMapper, s *stepWorker, d *tes.Executor) error {

	// Find the path for task stdin
	if d.Stdin != "" {
		stdin, err := mapper.OpenHostFile(d.Stdin)
		s.Command.SetStdin(stdin)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		stdout, err := mapper.CreateHostFile(d.Stdout)
		s.Command.SetStdout(stdout)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		stderr, err := mapper.CreateHostFile(d.Stderr)
		s.Command.SetStderr(stderr)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	return nil
}

// Validate the downloads/uploads.
func (r *DefaultWorker) validate(mapper *FileMapper) error {
	// TODO need to switch on directory type and check list as well.
	for _, input := range mapper.Inputs {
		unsupported := r.Store.UnsupportedOperations(input.Url)
		if unsupported.Get != nil {
			return fmt.Errorf("Input download not supported by storage: %v", unsupported.Get)
		}
	}
	for _, output := range mapper.Outputs {
		unsupported := r.Store.UnsupportedOperations(output.Url)
		if unsupported.Put != nil {
			return fmt.Errorf("Output upload not supported by storage: %v", unsupported.Put)
		}
	}
	return nil
}

func (r *DefaultWorker) pollForCancel(pctx context.Context, cancelCallback func()) context.Context {
	taskctx, cancel := context.WithCancel(pctx)

	// Start a goroutine that polls the server to watch for a canceled state.
	// If a cancel state is found, "taskctx" is canceled.
	go func() {
		ticker := time.NewTicker(time.Duration(r.Conf.PollingRate))
		defer ticker.Stop()

		for {
			select {
			case <-taskctx.Done():
				return
			case <-ticker.C:
				state, _ := r.TaskReader.State(taskctx)
				if tes.TerminalState(state) {
					cancel()
					cancelCallback()
				}
			}
		}
	}()
	return taskctx
}
