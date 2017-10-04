package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/version"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/rpc"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"path"
	"path/filepath"
	"time"
)

// NewDefaultWorker returns a new configured DefaultWorker instance.
func NewDefaultWorker(conf config.Worker) (Worker, error) {

	ev, err := events.FromConfig(conf.EventWriters)
	if err != nil {
		return nil, err
	}

	return &DefaultWorker{
		Conf:  conf,
		Event: ev,
	}, nil
}

// DefaultWorker is the default task worker, which follows a basic,
// sequential process of task initialization, execution, finalization,
// and logging.
type DefaultWorker struct {
	Conf  config.Worker
	Event events.Writer
}

// Run runs the Worker.
// TODO document behavior of slow consumer of task log updates
func (r *DefaultWorker) Run(ctx context.Context, task *tes.Task) {

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
	run.ctx = ctx

	ev := events.NewTaskWriter(task.Id, 0, r.Conf.Logger.Level, r.Event)
	ev.Info("Version", version.LogFields()...)

	// Map files into this baseDir
	baseDir := path.Join(r.Conf.WorkDir, task.Id)

	err := util.EnsureDir(baseDir)
	if err != nil {
		run.syserr = fmt.Errorf("failed to create worker baseDir: %v", err)
	}

	mapper := NewFileMapper(baseDir)

	if run.ok() {
		ev.State(tes.State_INITIALIZING)
	}

	ev.StartTime(time.Now())
	// Run the final logging/state steps in a deferred function
	// to ensure they always run, even if there's a missed error.
	defer func() {
		ev.EndTime(time.Now())

		switch {
		case run.execerr != nil:
			// One of the executors failed
			ev.Error("Exec error", run.execerr)
			ev.State(tes.State_ERROR)
		case run.syserr != nil:
			// Something else failed
			ev.Error("System error", run.syserr)
			ev.State(tes.State_SYSTEM_ERROR)
		default:
			ev.State(tes.State_COMPLETE)
		}
	}()

	// Recover from panics
	defer handlePanic(func(e error) {
		run.syserr = e
	})

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
		run.syserr = mapper.MapTask(task)
	}

	// Grab the IP address of this host. Used to send task metadata updates.
	var ip string
	if run.ok() {
		ip, run.syserr = externalIP()
	}

	// Configure a task-specific storage backend.
	// This provides download/upload for inputs/outputs.
	store := storage.Storage{}
	if run.ok() {
		store, run.syserr = store.WithConfig(r.Conf.Storage)
	}

	if run.ok() {
		run.syserr = r.validateInputs(&store, mapper)
	}

	if run.ok() {
		run.syserr = r.validateOutputs(&store, mapper)
	}

	// Download inputs
	for _, input := range mapper.Inputs {
		if run.ok() {
			run.syserr = store.Get(ctx, input.Url, input.Path, input.Type)
		}
	}

	if run.ok() {
		ev.State(tes.State_RUNNING)
	}

	// Run steps
	for i, d := range task.Executors {
		s := &stepWorker{
			Conf:  r.Conf,
			Event: ev.NewExecutorWriter(uint32(i)),
			IP:    ip,
			Cmd: &DockerCmd{
				ImageName:     d.ImageName,
				Cmd:           d.Cmd,
				Environ:       d.Environ,
				Volumes:       mapper.Volumes,
				Workdir:       d.Workdir,
				Ports:         d.Ports,
				ContainerName: fmt.Sprintf("%s-%d", task.Id, i),
				// TODO make RemoveContainer configurable
				RemoveContainer: true,
				Event:           ev.NewExecutorWriter(uint32(i)),
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

	// Upload outputs
	var outputs []*tes.OutputFileLog
	for _, output := range mapper.Outputs {
		if run.ok() {
			r.fixLinks(mapper, output.Path)
			var out []*tes.OutputFileLog
			out, run.syserr = store.Put(ctx, output.Url, output.Path, output.Type)
			outputs = append(outputs, out...)
		}
	}

	if run.ok() {
		ev.Outputs(outputs)
	}
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
func (r *DefaultWorker) openStepLogs(m *FileMapper, s *stepWorker, d *tes.Executor) error {

	// Find the path for task stdin
	var err error
	if d.Stdin != "" {
		s.Cmd.Stdin, err = m.OpenHostFile(d.Stdin)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Cmd.Stdout, err = m.CreateHostFile(d.Stdout)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Cmd.Stderr, err = m.CreateHostFile(d.Stderr)
		if err != nil {
			s.Event.Error("Couldn't prepare log files", err)
			return err
		}
	}
	return nil
}

// Validate the input downloads
func (r *DefaultWorker) validateInputs(store *storage.Storage, mapper *FileMapper) error {
	for _, input := range mapper.Inputs {
		if !store.Supports(input.Url, input.Path, input.Type) {
			return fmt.Errorf("Input download not supported by storage: %v", input)
		}
	}
	return nil
}

// Validate the output uploads
func (r *DefaultWorker) validateOutputs(store *storage.Storage, mapper *FileMapper) error {
	for _, output := range mapper.Outputs {
		if !store.Supports(output.Url, output.Path, output.Type) {
			return fmt.Errorf("Output upload not supported by storage: %v", output)
		}
	}
	return nil
}

func PollForCancel(ctx context.Context, id string, c *rpc.TESClient) context.Context {
	taskctx, cancel := context.WithCancel(ctx)

	// Start a goroutine that polls the server to watch for a canceled state.
	// If a cancel state is found, "taskctx" is canceled.
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-taskctx.Done():
				return
			case <-ticker.C:
				state, _ := c.State(id)
				if tes.TerminalState(state) {
					cancel()
				}
			}
		}
	}()
	return taskctx
}
