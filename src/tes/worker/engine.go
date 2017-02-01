package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"reflect"
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

// RunJob is a wrapper for runJob that polls for Cancel requests
// TODO documentation
func (eng *engine) RunJob(parentCtx context.Context, jobR *pbr.JobResponse) error {
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

	jobID := &pbe.JobID{
		Value: jobR.Job.JobID,
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	joberr := make(chan error, 1)
	go func() {
		joberr <- eng.runJob(ctx, sched, jobR)
	}()

	// Ticker for State polling
	tickChan := time.NewTicker(time.Millisecond * 10).C

	for {
		select {
		case joberr := <-joberr:
			if joberr != nil {
				sched.SetFailed(ctx, jobR.Job)
				return fmt.Errorf("Error running job: %v", joberr)
			}
			sched.SetComplete(ctx, jobR.Job)
			return nil
		case <-tickChan:
			jobDesc, err := sched.GetJobState(ctx, jobID)
			if err != nil {
				return fmt.Errorf("Error trying to get job status: %v", err)
			}
			switch jobDesc.State {
			case pbe.State_Canceled:
				cancel()
			}
		}
	}
}

// runJob runs a job
// TODO documentation
func (eng *engine) runJob(ctx context.Context, sched *scheduler.Client, jobR *pbr.JobResponse) error {
	// Initialize job
	sched.SetInitializing(ctx, jobR.Job)
	mapper, merr := eng.getMapper(jobR.Job)
	if merr != nil {
		sched.SetFailed(ctx, jobR.Job)
		return fmt.Errorf("Error during mapper initialization: %s", merr)
	}

	store, serr := eng.getStorage(jobR)
	if serr != nil {
		return fmt.Errorf("Error during store initialization: %s", serr)
	}

	derr := eng.downloadInputs(mapper, store)
	if derr != nil {
		return fmt.Errorf("Error during input provisioning: %s", derr)
	}

	// Run job steps
	sched.SetRunning(ctx, jobR.Job)
	for stepNum, step := range jobR.Job.Task.Docker {
		joberr := eng.runStep(ctx, sched, mapper, jobR.Job.JobID, step, stepNum)
		if joberr != nil {
			return fmt.Errorf("Error running job: %s", joberr)
		}
	}

	// Finalize job
	oerr := eng.uploadOutputs(mapper, store)
	if oerr != nil {
		return fmt.Errorf("Error uploading job outputs: %s", oerr)
	}

	// Job is Complete
	log.Println("Job completed without error")
	return nil
}

// runStep
// TODO documentation
func (eng *engine) runStep(ctx context.Context, sched *scheduler.Client, mapper *FileMapper, id string, step *pbe.DockerExecutor, stepNum int) error {
	stepID := fmt.Sprintf("%v-%v", id, stepNum)
	dcmd, err := eng.setupDockerCmd(mapper, step, stepID)
	if err != nil {
		return fmt.Errorf("Error setting up docker command: %v", err)
	}
	log.Printf("Running command: %s", strings.Join(dcmd.Cmd.Args, " "))

	// Start task step asynchronously
	dcmd.Cmd.Start()

	// Open channel to track async process
	done := make(chan error, 1)
	go func() {
		done <- dcmd.Cmd.Wait()
	}()

	// Open channel to track container initialization
	metaCh := make(chan []*pbe.Ports, 1)
	go func() {
		metaCh <- dcmd.InspectContainer(ctx)
	}()

	// Initialized to allow for DeepEquals comparison during polling
	stepLog := &pbe.JobLog{}

	// Ticker for polling rate
	tickChan := time.NewTicker(time.Millisecond * 10).C

	for {
		select {

		// ensure containers are stopped if the context is canceled
		// handles cancel request
		case <-ctx.Done():
			err := dcmd.StopContainer()
			if err != nil {
				return err
			}

		// TODO ensure metadata gets logged
		case portMap := <-metaCh:
			ip, err := externalIP()
			if err != nil {
				return err
			}

			// log update with host ip and port mapping
			initLog := &pbe.JobLog{
				HostIP: ip,
				Ports:  portMap,
			}

			statusReq := &pbr.UpdateStatusRequest{
				Id:   id,
				Step: int64(stepNum),
				Log:  initLog,
			}
			sched.UpdateJobStatus(ctx, statusReq)

		// handles docker run failure and success
		case cmdErr := <-done:
			stepLogUpdate := eng.finalizeLogs(dcmd, cmdErr)
			// final log update that includes the exit code
			statusReq := &pbr.UpdateStatusRequest{
				Id:   id,
				Step: int64(stepNum),
				Log:  stepLogUpdate,
			}
			sched.UpdateJobStatus(ctx, statusReq)

			if cmdErr != nil {
				return fmt.Errorf("Docker command error: %v", cmdErr)
			}
			return nil

		// update stdout and stderr in logs every 5 seconds
		case <-tickChan:
			stepLogUpdate := eng.updateLogs(dcmd)
			// check if log update has any new data
			if reflect.DeepEqual(stepLogUpdate, stepLog) == false {
				statusReq := &pbr.UpdateStatusRequest{
					Id:   id,
					Step: int64(stepNum),
					Log:  stepLogUpdate,
				}
				sched.UpdateJobStatus(ctx, statusReq)
			}
		}
	}
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
		ImageName:     step.ImageName,
		CmdString:     step.Cmd,
		Volumes:       mapper.Volumes,
		Workdir:       step.Workdir,
		Ports:         step.Ports,
		ContainerName: id,
		// TODO make RemoveContainer configurable
		RemoveContainer: true,
		Stdin:           nil,
		Stdout:          nil,
		Stderr:          nil,
		Log:             map[string][]byte{},
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

	dcmd, err := dcmd.SetupCommand()
	if err != nil {
		return nil, fmt.Errorf("Error setting up job command: %s", err)
	}

	return dcmd, nil
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down

		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface

		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err

		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("Error no network connection")
}

func (eng *engine) updateLogs(dcmd *DockerCmd) *pbe.JobLog {
	stepLog := &pbe.JobLog{}

	if len(dcmd.Log["Stdout"]) > 0 {
		stdoutText := string(dcmd.Log["Stdout"][:])
		dcmd.Log["Stdout"] = []byte{}
		stepLog.Stdout = stdoutText
	}

	if len(dcmd.Log["Stderr"]) > 0 {
		stderrText := string(dcmd.Log["Stderr"][:])
		dcmd.Log["Stderr"] = []byte{}
		stepLog.Stderr = stderrText
	}

	return stepLog
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
