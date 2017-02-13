package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	pbr "tes/server/proto"
	"tes/storage"
)

// TODO document behavior of slow consumer of updates
func runJob(ctx context.Context, resp *pbr.JobResponse, conf config.Worker, up updateChan) error {
	job := resp.Job

	// Create working dir
	dir, err := filepath.Abs(conf.WorkDir)
	if err != nil {
		return err
	}
	ensureDir(dir)

	// Map task paths to working dir paths
	mapper := NewJobFileMapper(job.JobID, conf.WorkDir)
	merr := mapper.MapTask(job.Task)
	if merr != nil {
		return merr
	}

	// Configure a job-specific storage backend.
	// This provides download/upload for inputs/outputs.
	store, serr := getStorage(conf.Storage)
	if serr != nil {
		return fmt.Errorf("Error during store initialization: %s", serr)
	}

	// Grab the IP address of this host. Used to send job metadata updates.
	ip, ierr := externalIP()
	if ierr != nil {
		return ierr
	}

	r := &jobRunner{
		conf,
		job,
		job.Task.Docker,
		mapper,
		store,
		ip,
		up,
		log.WithFields("jobID", job.JobID),
	}
	return r.Run(ctx)
}

type jobRunner struct {
	conf    config.Worker
	job     *pbe.Job
	steps   []*pbe.DockerExecutor
	mapper  *FileMapper
	store   *storage.Storage
	ip      string
	updates updateChan
	log     logger.Logger
}

func (r *jobRunner) Run(ctx context.Context) error {
	r.log.Debug("JobRunner.Run", "jobID", r.job.JobID)
	// The code here is verbose, but simple; mainly loops and simple error checking.
	//
	// The steps are:
	// 1. validate input and output mappings
	// 2. download inputs
	// 3. run the steps (docker)
	// 4. upload the outputs
	//
	// If any part fails, processing stops and an error is returned.
	// If the context "ctx" is canceled, processing stops.
	//
	// Errors and context have to be checked frequently, hence the verbose code.

	// helps clean up the frequent context checking in the loops below.
	isCtxDone := func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	// Validate the input downloads
	/* TODO Supports() needs to be added to the Storage interface
		for _, input := range r.mapper.Inputs {
	    if !r.store.Supports(in.Location, in.Path, in.Class) {
	      return fmt.Error("Input download not supported by storage: %v", in)
	    }
		}

	  // Validate the output uploads
		for _, output := range mapper.Outputs {
	    if !r.store.Supports(out.Location, out.Path, out.Class) {
	      return fmt.Error("Output upload not supported by storage: %v", out)
	    }
		}
	*/

	// Validate and prepare the step commands
	stepRunners := make([]*stepRunner, 0, len(r.steps))
	for i, step := range r.steps {
		s := &stepRunner{
			JobID:   r.job.JobID,
			Conf:    r.conf,
			Num:     i,
			Log:     r.log.WithFields("step", i),
			Updates: r.updates,
			IP:      r.ip,
			Cmd: &DockerCmd{
				ImageName:     step.ImageName,
				Cmd:           step.Cmd,
				Volumes:       r.mapper.Volumes,
				Workdir:       step.Workdir,
				Ports:         step.Ports,
				ContainerName: fmt.Sprintf("%s-%d", r.job.JobID, i),
				// TODO make RemoveContainer configurable
				RemoveContainer: true,
			},
		}

		// Opens stdin/out/err files and updates those fields on "cmd".
		err := r.openStepLogs(s, step)
		if err != nil {
			s.Log.Error("Couldn't prepare log files", err)
			return err
		}

		stepRunners = append(stepRunners, s)
	}

	// Download inputs
	for _, input := range r.mapper.Inputs {
		if isCtxDone() {
			return ctx.Err()
		}

		// Storage Get
		// TODO should be passed context so it can cancel
		err := r.store.Get(input.Location, input.Path, input.Class)

		if err != nil {
			return err
		}
	}

	// Run steps
	for _, s := range stepRunners {
		if isCtxDone() {
			return ctx.Err()
		}

		err := s.Run(ctx)

		if err != nil {
			return err
		}
	}

	// Upload outputs
	for _, output := range r.mapper.Outputs {
		if isCtxDone() {
			return ctx.Err()
		}

		// Storage Put
		// TODO should be passed context so it can cancel
		err := r.store.Put(output.Location, output.Path, output.Class)

		if err != nil {
			return err
		}
	}

	return nil
}

// openLogs opens/creates the logs files for a step and updates those fields.
func (r *jobRunner) openStepLogs(s *stepRunner, step *pbe.DockerExecutor) error {

	// Find the path for job stdin
	var err error
	if step.Stdin != "" {
		s.Cmd.Stdin, err = r.mapper.OpenHostFile(step.Stdin)
		if err != nil {
			return err
		}
	}

	// Create file for job stdout
	if step.Stdout != "" {
		s.Cmd.Stdout, err = r.mapper.CreateHostFile(step.Stdout)
		if err != nil {
			return err
		}
	}

	// Create file for job stderr
	if step.Stderr != "" {
		s.Cmd.Stderr, err = r.mapper.CreateHostFile(step.Stderr)
		if err != nil {
			return err
		}
	}
	return nil
}

// getStorage returns a Storage instance configured for the given job.
func getStorage(confs []*config.StorageConfig) (*storage.Storage, error) {
	var err error
	storage := new(storage.Storage)

	for _, conf := range confs {
		storage, err = storage.WithConfig(conf)
		if err != nil {
			return nil, err
		}
	}

	if storage == nil {
		return nil, fmt.Errorf("No storage configured")
	}

	return storage, nil
}
