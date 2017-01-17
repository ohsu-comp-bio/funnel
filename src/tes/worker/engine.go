package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	pbe "tes/ga4gh"
	"tes/scheduler"
	pbr "tes/server/proto"
	"tes/storage"
	"time"
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
	sched.SetRunning(ctx, job.Job)
	joberr := eng.runJob(ctx, sched, job)

	// Tell the scheduler if the job failed
	if joberr != nil {
		sched.SetFailed(ctx, jobR.Job)
		log.Printf("Failed to run job: %s", jobR.Job.JobID)
		log.Printf("%s", joberr)
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
		stepID := fmt.Sprintf("%v-%v", job.JobID, stepNum)
		dcmd, err := eng.setupDockerCmd(mapper, step, stepID)
		if err != nil {
			return fmt.Errorf("Error setting up docker command: %v", err)
		}
		log.Printf("Running command: %s", strings.Join(dcmd.ExecCmd.Args, " "))
		// Start task step asynchronously
		dcmd.ExecCmd.Start()

		// Get container metadata
		metadata, mq_err := dcmd.GetContainerMetadata()
		if mq_err != nil {
			return fmt.Errorf("queryContainerMetadata Error: %s", mq_err)
		}
		log.Printf("Metadata: %s", metadata)
		// Send the scheduler service an initial job status update that includes
		// the metadata
		statusReq := &pbr.UpdateStatusRequest{
			Id:       job.JobID,
			Step:     int64(stepNum),
			Metadata: metadata,
		}
		sched.UpdateJobStatus(ctx, statusReq)

		// Open chanel to track async process
		done := make(chan error, 1)
		go func() {
			done <- dcmd.ExecCmd.Wait()
		}()

		jobID := &pbe.JobID{
			Value: job.JobID,
		}

		// Ticker for polling rate
		// TODO figure out a reasonable rate
		tickChan := time.NewTicker(time.Second * 5).C
	PollLoop:
		for {
			select {
			case cmd_err := <-done:
				stepLog := eng.finalizeLogs(dcmd, cmd_err)
				// Send the scheduler service a final job status update that includes
				// the exit code
				statusReq := &pbr.UpdateStatusRequest{
					Id:   job.JobID,
					Step: int64(stepNum),
					Log:  stepLog,
				}
				sched.UpdateJobStatus(ctx, statusReq)

				if cmd_err != nil {
					return fmt.Errorf("Docker cmd error: %v", cmd_err)
				} else {
					log.Print("Process finished gracefully without error")
					sched.SetComplete(ctx, job)
					break PollLoop
				}
			case <-tickChan:
				jobDesc, err := sched.GetJobStatus(ctx, jobID)
				if err != nil {
					return fmt.Errorf("Error trying to get job status: %v", err)
				}
				switch jobDesc.State {
				case pbe.State_Canceled:
					err := dcmd.StopContainer()
					if err != nil {
						return fmt.Errorf("Error trying to stop container: %v", err)
					} else {
						log.Print("Successfully canceled job")
						break PollLoop
					}
				case pbe.State_Running:
					log.Print("Waiting for task to finish...")
					stepLog := eng.updateLogs(dcmd)
					// Send the scheduler service a job status update with updates
					// to stdout and stderr
					statusReq := &pbr.UpdateStatusRequest{
						Id:   job.JobID,
						Step: int64(stepNum),
						Log:  stepLog,
					}
					sched.UpdateJobStatus(ctx, statusReq)
					continue
				default:
					return fmt.Errorf("Job improperly acquired status: %v", jobDesc.State)
				}
			}
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
func (eng *engine) setupDockerCmd(mapper *FileMapper, step *pbe.DockerExecutor, id string) (*DockerCmd, error) {

	dcmd := &DockerCmd{
		ImageName: step.ImageName,
		Cmd:       step.Cmd,
		Volumes:   mapper.Volumes,
		Workdir:   step.Workdir,
		// TODO make Port configurable
		Port:          "8888:8888",
		ContainerName: id,
		// TODO make RemoveContainer configurable
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
		dcmd.Stdin = f
	}

	// Create file for job stdout
	if step.Stdout != "" {
		f, err := mapper.CreateHostFile(step.Stdout)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stdout: %s", err)
		}
		dcmd.Stdout = f
	}

	// Create file for job stderr
	if step.Stderr != "" {
		f, err := mapper.CreateHostFile(step.Stderr)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stderr: %s", err)
		}
		dcmd.Stderr = f
	}

	return dcmd.SetupCommand(), nil
}

func (eng *engine) updateLogs(dcmd *DockerCmd) *pbe.JobLog {
	// TODO rethink these messages. You probably don't want head().
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
		Stdout: stdoutText,
		Stderr: stderrText,
	}

	return steplog
}

func (eng *engine) finalizeLogs(dcmd *DockerCmd, cmdErr error) *pbe.JobLog {
	exitCode := getExitCode(cmdErr)
	log.Printf("Exit code: %d", exitCode)
	steplog := eng.updateLogs(dcmd)
	steplog.ExitCode = exitCode
	return steplog
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
