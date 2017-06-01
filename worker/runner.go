package worker

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"path"
	"path/filepath"
	"time"
)

// TaskRunner is a function that does the work of running a task on a worker,
// including download inputs, executing commands, uploading outputs, etc.
type TaskRunner func(TaskControl, config.Worker, *pbf.TaskWrapper)

// Default TaskRunner
func runTask(ctrl TaskControl, conf config.Worker, t *pbf.TaskWrapper) {
	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, t.Task.Id)
	client, _ := newSchedClient(conf)

	r := &taskRunner{
		ctrl:    ctrl,
		wrapper: t,
		mapper:  NewFileMapper(baseDir),
		store:   storage.Storage{},
		conf:    conf,
		taskLogger: &RPCTask{
			client: client,
			taskID: t.Task.Id,
		},
		log: logger.New("runner", "workerID", conf.ID, "taskID", t.Task.Id),
	}
	go r.Run()
}

// taskRunner helps collect data used across many helper methods.
type taskRunner struct {
	ctrl       TaskControl
	wrapper    *pbf.TaskWrapper
	conf       config.Worker
	taskLogger TaskLogger
	log        logger.Logger
	mapper     *FileMapper
	store      storage.Storage
	ip         string
}

// TODO document behavior of slow consumer of task log updates
func (r *taskRunner) Run() {
	r.log.Debug("TaskRunner.Run")
	task := r.wrapper.Task
	// The code here is verbose, but simple; mainly loops and simple error checking.
	//
	// The steps are:
	// 1. validate input and output mappings
	// 2. download inputs
	// 3. run the steps (docker)
	// 4. upload the outputs
	r.taskLogger.StartTime(time.Now())

	r.step("prepareDir", r.prepareDir)
	r.step("prepareMapper", r.prepareMapper)
	r.step("prepareStorage", r.prepareStorage)
	r.step("prepareIP", r.prepareIP)
	r.step("validateInputs", r.validateInputs)
	r.step("validateOutputs", r.validateOutputs)

	// Download inputs
	for _, input := range r.mapper.Inputs {
		r.step("store.Get", func() error {
			return r.store.Get(
				r.ctrl.Context(),
				input.Url,
				input.Path,
				input.Type,
			)
		})
	}

	r.ctrl.SetRunning()

	// Run steps
	for i, d := range task.Executors {
		stepName := fmt.Sprintf("step-%d", i)
		r.step(stepName, func() error {
			s := &stepRunner{
				TaskID:     task.Id,
				Conf:       r.conf,
				Num:        i,
				Log:        r.log.WithFields("step", i),
				TaskLogger: r.taskLogger,
				IP:         r.ip,
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
			err := r.openStepLogs(s, d)
			if err != nil {
				s.Log.Error("Couldn't prepare log files", err)
				return err
			}
			return s.Run(r.ctrl.Context())
		})
	}

	// Upload outputs
	log.Debug("Outputs", r.mapper.Outputs)
	var outputs []*tes.OutputFileLog
	for _, output := range r.mapper.Outputs {
		r.step("store.Put", func() error {
			r.fixLinks(output.Path)
			out, err := r.store.Put(r.ctrl.Context(), output.Url, output.Path, output.Type)
			outputs = append(outputs, out...)
			return err
		})
	}

	r.taskLogger.Outputs(outputs)
	r.taskLogger.EndTime(time.Now())

	r.ctrl.SetResult(nil)
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
			return err
		}
	}

	// Create file for task stdout
	if d.Stdout != "" {
		s.Cmd.Stdout, err = r.mapper.CreateHostFile(d.Stdout)
		if err != nil {
			return err
		}
	}

	// Create file for task stderr
	if d.Stderr != "" {
		s.Cmd.Stderr, err = r.mapper.CreateHostFile(d.Stderr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create working dir
func (r *taskRunner) prepareDir() error {
	dir, err := filepath.Abs(r.conf.WorkDir)
	if err != nil {
		return err
	}
	return util.EnsureDir(dir)
}

// Prepare file mapper, which maps task file URLs to host filesystem paths
func (r *taskRunner) prepareMapper() error {
	// Map task paths to working dir paths
	return r.mapper.MapTask(r.wrapper.Task)
}

// Grab the IP address of this host. Used to send task metadata updates.
func (r *taskRunner) prepareIP() error {
	var err error
	r.ip, err = externalIP()
	return err
}

// Configure a task-specific storage backend.
// This provides download/upload for inputs/outputs.
func (r *taskRunner) prepareStorage() error {
	var err error

	r.store, err = r.store.WithConfig(r.conf.Storage)
	if err != nil {
		return err
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

// step helps clean up the frequent context and error checking code.
//
// Every operation in the runner needs to check if the context is done,
// and handle errors appropriately. This helper removes that duplicated, verbose code.
func (r *taskRunner) step(name string, stepfunc func() error) {
	// If the runner is already complete (perhaps because a previous step failed)
	// skip the step.
	if !r.ctrl.Complete() {
		// Run the step
		err := stepfunc()
		// If the step failed, set the runner to failed. All the following steps
		// will be skipped.
		if err != nil {
			r.log.Error("Task runner step failed", "error", err, "step", name)
			r.ctrl.SetResult(err)
		}
	}
}
