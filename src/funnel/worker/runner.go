package worker

import (
	"fmt"
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"funnel/storage"
	"funnel/util"
	"os"
	"path"
	"path/filepath"
)

// JobRunner is a function that does the work of running a job on a worker,
// including download inputs, executing commands, uploading outputs, etc.
type JobRunner func(JobControl, config.Worker, *pbf.JobWrapper, logUpdateChan)

// Default JobRunner
func runJob(ctrl JobControl, conf config.Worker, j *pbf.JobWrapper, up logUpdateChan) {
	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, j.Job.JobID)

	r := &jobRunner{
		ctrl:    ctrl,
		wrapper: j,
		mapper:  NewFileMapper(baseDir),
		store:   &storage.Storage{},
		conf:    conf,
		updates: up,
		log:     logger.New("runner", "workerID", conf.ID, "jobID", j.Job.JobID),
	}
	go r.Run()
}

// jobRunner helps collect data used across many helper methods.
type jobRunner struct {
	ctrl    JobControl
	wrapper *pbf.JobWrapper
	conf    config.Worker
	updates logUpdateChan
	log     logger.Logger
	mapper  *FileMapper
	store   *storage.Storage
	ip      string
}

// TODO document behavior of slow consumer of job log updates
func (r *jobRunner) Run() {
	r.log.Debug("JobRunner.Run")
	job := r.wrapper.Job
	// The code here is verbose, but simple; mainly loops and simple error checking.
	//
	// The steps are:
	// 1. validate input and output mappings
	// 2. download inputs
	// 3. run the steps (docker)
	// 4. upload the outputs

	r.step("prepareDir", r.prepareDir)
	r.step("prepareMapper", r.prepareMapper)
	r.step("prepareStorage", r.prepareStorage)
	r.step("prepareIP", r.prepareIP)
	r.step("validateInputs", r.validateInputs)
	r.step("validateOutputs", r.validateOutputs)

	// Download inputs
	for _, input := range r.mapper.Inputs {
		r.step("store.Get", func() error {
			vol, _ := r.mapper.FindVolume(input.Path)
			return r.store.Get(
				r.ctrl.Context(),
				input.Location,
				input.Path,
				input.Class,
				vol.Readonly,
			)
		})
	}

	r.ctrl.SetRunning()

	// Run steps
	for i, d := range job.Task.Docker {
		stepName := fmt.Sprintf("step-%d", i)
		r.step(stepName, func() error {
			s := &stepRunner{
				JobID:   job.JobID,
				Conf:    r.conf,
				Num:     i,
				Log:     r.log.WithFields("step", i),
				Updates: r.updates,
				IP:      r.ip,
				Cmd: &DockerCmd{
					ImageName:     d.ImageName,
					Cmd:           d.Cmd,
					Volumes:       r.mapper.Volumes,
					Workdir:       d.Workdir,
					Ports:         d.Ports,
					ContainerName: fmt.Sprintf("%s-%d", job.JobID, i),
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
	for _, output := range r.mapper.Outputs {
		r.step("store.Put", func() error {
			r.fixLinks(output.Path)
			return r.store.Put(r.ctrl.Context(), output.Location, output.Path, output.Class)
		})
	}

	r.ctrl.SetResult(nil)
}

// fixLinks walks the output paths, fixing cases where a symlink is
// broken because it's pointing to a path inside a container volume.
func (r *jobRunner) fixLinks(basepath string) {
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
func (r *jobRunner) openStepLogs(s *stepRunner, d *tes.DockerExecutor) error {

	// Find the path for job stdin
	var err error
	if d.Stdin != "" {
		s.Cmd.Stdin, err = r.mapper.OpenHostFile(d.Stdin)
		if err != nil {
			return err
		}
	}

	// Create file for job stdout
	if d.Stdout != "" {
		s.Cmd.Stdout, err = r.mapper.CreateHostFile(d.Stdout)
		if err != nil {
			return err
		}
	}

	// Create file for job stderr
	if d.Stderr != "" {
		s.Cmd.Stderr, err = r.mapper.CreateHostFile(d.Stderr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create working dir
func (r *jobRunner) prepareDir() error {
	dir, err := filepath.Abs(r.conf.WorkDir)
	if err != nil {
		return err
	}
	return util.EnsureDir(dir)
}

// Prepare file mapper, which maps task file URLs to host filesystem paths
func (r *jobRunner) prepareMapper() error {
	// Map task paths to working dir paths
	return r.mapper.MapTask(r.wrapper.Job.Task)
}

// Grab the IP address of this host. Used to send job metadata updates.
func (r *jobRunner) prepareIP() error {
	var err error
	r.ip, err = externalIP()
	return err
}

// Configure a job-specific storage backend.
// This provides download/upload for inputs/outputs.
func (r *jobRunner) prepareStorage() error {
	var err error

	for _, conf := range r.conf.Storage {
		r.store, err = r.store.WithConfig(conf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate the input downloads
func (r *jobRunner) validateInputs() error {
	for _, input := range r.mapper.Inputs {
		if !r.store.Supports(input.Location, input.Path, input.Class) {
			return fmt.Errorf("Input download not supported by storage: %v", input)
		}
	}
	return nil
}

// Validate the output uploads
func (r *jobRunner) validateOutputs() error {
	for _, output := range r.mapper.Outputs {
		if !r.store.Supports(output.Location, output.Path, output.Class) {
			return fmt.Errorf("Output upload not supported by storage: %v", output)
		}
	}
	return nil
}

// step helps clean up the frequent context and error checking code.
//
// Every operation in the runner needs to check if the context is done,
// and handle errors appropriately. This helper removes that duplicated, verbose code.
func (r *jobRunner) step(name string, stepfunc func() error) {
	// If the runner is already complete (perhaps because a previous step failed)
	// skip the step.
	if !r.ctrl.Complete() {
		// Run the step
		err := stepfunc()
		// If the step failed, set the runner to failed. All the following steps
		// will be skipped.
		if err != nil {
			r.log.Error("Job runner step failed", "error", err, "step", name)
			r.ctrl.SetResult(err)
		}
	}
}
