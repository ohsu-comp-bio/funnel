package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	pbr "tes/server/proto"
	"tes/storage"
)

func newJobRunner(conf config.Worker, a *pbr.Assignment, up updateChan) *jobRunner {
	return &jobRunner{
		Assignment: a,
		mapper:     NewJobFileMapper(a.Job.JobID, conf.WorkDir),
		store:      &storage.Storage{},
		conf:       conf,
		updates:    up,
		log:        logger.New("runner", "workerID", conf.ID, "jobID", a.Job.JobID),
	}
}

type jobRunner struct {
	Job        *pbe.Job
	Assignment *pbr.Assignment
	conf       config.Worker
	updates    updateChan
	log        logger.Logger
	mapper     *FileMapper
	store      *storage.Storage
	ip         string
	cancelFunc context.CancelFunc
	running    bool
	complete   bool
	err        error
	mtx        sync.Mutex
}

// TODO document behavior of slow consumer of updates
func (r *jobRunner) Run() {
	r.log.Debug("JobRunner.Run", "jobID", r.Assignment.Job.JobID)
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFunc = cancel
	job := r.Assignment.Job
	// The code here is verbose, but simple; mainly loops and simple error checking.
	//
	// The steps are:
	// 1. validate input and output mappings
	// 2. download inputs
	// 3. run the steps (docker)
	// 4. upload the outputs

	r.step(ctx, r.prepareDir)
	r.step(ctx, r.prepareMapper)
	r.step(ctx, r.prepareStorage)
	r.step(ctx, r.prepareIP)
	r.step(ctx, r.validateInputs)
	r.step(ctx, r.validateOutputs)

	// Download inputs
	for _, input := range r.mapper.Inputs {
		r.step(ctx, func() error {
			return r.store.Get(ctx, input.Location, input.Path, input.Class)
		})
	}

	r.setRunning()

	// Run steps
	for i, d := range job.Task.Docker {
		r.step(ctx, func() error {
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
			return s.Run(ctx)
		})
	}

	// Upload outputs
	for _, output := range r.mapper.Outputs {
		r.step(ctx, func() error {
			return r.store.Put(ctx, output.Location, output.Path, output.Class)
		})
	}

	r.setResult(nil)
}

// openLogs opens/creates the logs files for a step and updates those fields.
func (r *jobRunner) openStepLogs(s *stepRunner, d *pbe.DockerExecutor) error {

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
	return ensureDir(dir)
}

// Prepare file mapper, which maps task file URLs to host filesystem paths
func (r *jobRunner) prepareMapper() error {
	// Map task paths to working dir paths
	return r.mapper.MapTask(r.Assignment.Job.Task)
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
func (r *jobRunner) step(ctx context.Context, stepfunc func() error) {
	select {
	case <-ctx.Done():
		r.setResult(ctx.Err())
	default:
		// If the runner is already complete (perhaps because a previous step failed)
		// skip the step.
		if !r.Complete() {
			// Run the step
			err := stepfunc()
			// If the step failed, set the runner to failed. All the following steps
			// will be skipped.
			if err != nil {
				r.log.Error("Job runner failed", err)
				r.setResult(err)
			}
		}
	}
}

func (r *jobRunner) setResult(err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	// Don't set the result twice
	if !r.complete {
		r.complete = true
		r.err = err
	}
}

func (r *jobRunner) Err() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.err
}

func (r *jobRunner) Cancel() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.cancelFunc != nil {
		r.cancelFunc()
	}
}

func (r *jobRunner) Complete() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.complete
}

func (r *jobRunner) State() pbe.State {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	switch {
	case r.err == context.Canceled:
		return pbe.State_Canceled
	case r.err != nil:
		return pbe.State_Error
	case r.complete:
		return pbe.State_Complete
	case r.running:
		return pbe.State_Running
	default:
		return pbe.State_Initializing
	}
}

func (r *jobRunner) setRunning() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.complete {
		r.running = true
	}
}
