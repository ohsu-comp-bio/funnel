package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"syscall"
	pbe "tes/ga4gh"
	"tes/scheduler"
	pbr "tes/server/proto"
	"tes/storage"
)

// Engine is responsible for running a job. This includes downloading inputs,
// communicating updates to the scheduler service, running the actual command,
// and uploading outputs.
type Engine interface {
	RunJob(ctx context.Context, job *pbr.JobResponse) error
}

// engine is the internal implementation of a docker job engine.
type engine struct {
	conf Config
}

// NewEngine returns a new Engine instance configured with a given scheduler address,
// working directory, and storage client.
//
// If the working directory can't be initialized, this returns an error.
func NewEngine(conf Config) (Engine, error) {
	dir, err := filepath.Abs(conf.WorkDir)
	if err != nil {
		return nil, err
	}
	ensureDir(dir)

	return &engine{conf}, nil
}

// RunJob runs a job.
func (eng *engine) RunJob(ctx context.Context, jobR *pbr.JobResponse) error {
	// This is essentially a simple helper for runJob() (below).
	// This ensures that the job state is always updated in the scheduler,
	// without having to do it on 15+ different lines in runJob() and others.
	//
	// Please try to keep this function as simple as possible.
	// New code should probably go in runJob()

	// Get a client for the scheduler service
	sched, schederr := scheduler.NewClient(eng.conf.ServerAddress)
	defer sched.Close()
	// TODO if we're here then we have a serious problem. We have already
	//      told the scheduler that we're running the job, but now we can't
	//      tell it things are broken, so the job is going to stay running
	//      forever. Possibly the scheduler should have a job timeout.
	if schederr != nil {
		return schederr
	}

	// Tell the scheduler the job is running.
	sched.SetRunning(ctx, jobR.Job)
	joberr := eng.runJob(ctx, sched, jobR)

	// Tell the scheduler whether the job failed or completed.
	if joberr != nil {
		sched.SetFailed(ctx, jobR.Job)
		//BUG: error status not returned to scheduler
		log.Printf("Failed to run job: %s", jobR.Job.JobID)
		log.Printf("%s", joberr)
	} else {
		sched.SetComplete(ctx, jobR.Job)
	}
	return joberr
}

// runJob calls a series of other functions to process a job:
// 1. set up the file mapping between the host and the container
// 2. set up the storage client
// 3. download the inputs
// 4. run the job steps
// 4a. update the scheduler with job status after each step
// 5. upload the outputs
func (eng *engine) runJob(ctx context.Context, sched *scheduler.Client, jobR *pbr.JobResponse) error {
	job := jobR.Job
	mapper, merr := eng.getMapper(job)
	if merr != nil {
		return merr
	}

	// TODO catch error
	store, serr := eng.getStorage(jobR)
	if serr != nil {
		return serr
	}

	derr := eng.downloadInputs(mapper, store)
	if derr != nil {
		return derr
	}

	// TODO is it possible to allow context.Done() to kill the current step?
	for stepNum, step := range job.Task.Docker {
		stepLog, err := eng.runStep(mapper, step)

		if stepLog != nil {
			// Send the scheduler service a job status update
			statusReq := &pbr.UpdateStatusRequest{
				Id:   job.JobID,
				Step: int64(stepNum),
				Log:  stepLog,
			}
			sched.UpdateJobStatus(ctx, statusReq)
		}
		if err != nil {
			return err
		}
	}

	uerr := eng.uploadOutputs(mapper, store)
	if uerr != nil {
		return uerr
	}

	return nil
}

// getMapper returns a FileMapper instance with volumes, inputs, and outputs
// configured for the given job.
func (eng *engine) getMapper(job *pbe.Job) (*FileMapper, error) {
	mapper := NewJobFileMapper(job.JobID, eng.conf.WorkDir)

	// Iterates through job.Task.Resources.Volumes and add the volume to mapper.
	for _, vol := range job.Task.Resources.Volumes {
		err := mapper.AddVolume(vol.Source, vol.MountPoint)
		if err != nil {
			return nil, err
		}
	}

	// Add all the inputs to the mapper
	for _, input := range job.Task.Inputs {
		err := mapper.AddInput(input)
		if err != nil {
			return nil, err
		}
	}

	// Add all the outputs to the mapper
	for _, output := range job.Task.Outputs {
		err := mapper.AddOutput(output)
		if err != nil {
			return nil, err
		}
	}

	return mapper, nil
}

// getStorage returns a Storage instance configured for the given job.
func (eng *engine) getStorage(jobR *pbr.JobResponse) (*storage.Storage, error) {
	var err error
	storage := new(storage.Storage)

	for _, conf := range eng.conf.Storage {
		storage, err = storage.WithConfig(conf)
		if err != nil {
			return nil, err
		}
	}

	for _, conf := range jobR.Storage {
		storage, err = storage.WithConfig(conf)
		if err != nil {
			return nil, err
		}
	}

	return storage, nil
}

func (eng *engine) downloadInputs(mapper *FileMapper, store *storage.Storage) error {
	// Validate all the input source URLs
	//for _, input := range mapper.Inputs {
	// TODO ?
	//}

	// Download all the inputs from storage
	for _, input := range mapper.Inputs {
		err := store.Get(input.Location, input.Path, input.Class)
		if err != nil {
			return err
		}
	}
	return nil
}

// The bulk of job running happens here.
func (eng *engine) runStep(mapper *FileMapper, step *pbe.DockerExecutor) (*pbe.JobLog, error) {

	dcmd := DockerCmd{
		ImageName: step.ImageName,
		Cmd:       step.Cmd,
		Volumes:   mapper.Volumes,
		Workdir:   step.Workdir,
		// TODO make this configurable
		RemoveContainer: true,
		Stdin:           nil,
		Stdout:          nil,
		Stderr:          nil,
	}

	// Find the path for job stdin
	if step.Stdin != "" {
		f, err := mapper.OpenHostFile(step.Stdin)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stdin: %s", err)
		}
		defer f.Close()
		dcmd.Stdin = f
	}

	// Create file for job stdout
	if step.Stdout != "" {
		f, err := mapper.CreateHostFile(step.Stdout)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stdout: %s", err)
		}
		defer f.Close()
		dcmd.Stdout = f
	}

	// Create file for job stderr
	if step.Stderr != "" {
		f, err := mapper.CreateHostFile(step.Stderr)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stderr: %s", err)
		}
		defer f.Close()
		dcmd.Stderr = f
	}

	cmdErr := dcmd.Run()
	exitCode := getExitCode(cmdErr)
	log.Printf("Exit code: %d", exitCode)

	// TODO rethink these messages. You probably don't want head().
	//      you also don't get this until the step is finished,
	//      when you really want streaming.
	//
	// Get the head of the stdout/stderr files, if they exist.
	stdoutText := ""
	stderrText := ""
	if dcmd.Stdout != nil {
		stdoutText = readFileHead(dcmd.Stdout.Name())
	}
	if dcmd.Stderr != nil {
		stderrText = readFileHead(dcmd.Stderr.Name())
	}

	steplog := &pbe.JobLog{
		Stdout:   stdoutText,
		Stderr:   stderrText,
		ExitCode: exitCode,
	}

	if cmdErr != nil {
		return steplog, cmdErr
	}
	return steplog, nil
}

func (eng *engine) uploadOutputs(mapper *FileMapper, store *storage.Storage) error {
	// Upload all the outputs to storage
	for _, out := range mapper.Outputs {
		err := store.Put(out.Location, out.Path, out.Class)
		if err != nil {
			return err
		}
	}
	return nil
}

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitCode(err error) int32 {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return int32(status.ExitStatus())
			}
		} else {
			log.Printf("Could not determine exit code. Using default -999")
			return -999
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
}
