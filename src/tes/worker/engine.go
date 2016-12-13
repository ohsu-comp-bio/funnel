package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"tes/ga4gh"
	"tes/scheduler"
	"tes/server/proto"
)

// Used to pass file system configuration to the worker engine.
type FileConfig struct {
	SwiftCacheDir string
	AllowedDirs   string
	SharedDir     string
	VolumeDir     string
}

const headerSize = int64(102400)

func readFileHead(path string) []byte {
	f, _ := os.Open(path)
	buffer := make([]byte, headerSize)
	l, _ := f.Read(buffer)
	f.Close()
	return buffer[:l]
}

func FindHostPath(bindings []FSBinding, containerPath string) string {
	for _, binding := range bindings {
		if binding.ContainerPath == containerPath {
			return binding.HostPath
		}
	}
	return ""
}

func FindStdin(bindings []FSBinding, containerPath string) (*os.File, error) {
	stdinPath := FindHostPath(bindings, containerPath)
	if stdinPath != "" {
		return os.Open(stdinPath)
	} else {
		return nil, nil
	}
}

type Engine interface {
	RunJob(ctx context.Context, job *ga4gh_task_exec.Job) error
}

type engine struct {
	sched  *scheduler.Client
	mapper *FileMapper
}

func NewEngine(schedAddr string, fconf FileConfig) (Engine, error) {

	// Get a client for the scheduler service
	sched, err := scheduler.NewClient(schedAddr)
	if err != nil {
		return nil, err
	}

	fileClient := getFileClient(fconf)
	mapper := NewFileMapper(fileClient, fconf.VolumeDir)
	return &engine{sched, mapper}, nil
}

func (eng *engine) Close() {
	eng.sched.Close()
	// TODO does file mapper or anythign else need cleanup?
}

// RunJob runs a job.
func (eng *engine) RunJob(ctx context.Context, job *ga4gh_task_exec.Job) error {
	// Tell the scheduler the job is running
	eng.sched.SetRunning(ctx, job)
	err := eng.runJob(ctx, job)
	if err != nil {
		eng.sched.SetFailed(ctx, job)
		//BUG: error status not returned to scheduler
		log.Printf("Failed to run job [%s]: %s", job.JobID, err)
	} else {
		eng.sched.SetComplete(ctx, job)
	}
	return err
}

func (eng *engine) runJob(ctx context.Context, job *ga4gh_task_exec.Job) error {
	// Modifies the filemapper's jobID
	mapper := eng.mapper
	mapper.Job(job.JobID)

	// Iterates through job.Task.Resources.Volumes and add the volume to mapper.
	for _, disk := range job.Task.Resources.Volumes {
		mapper.AddVolume(job.JobID, disk.Source, disk.MountPoint)
	}

	// MapInput copies the input.Location into input.Path.
	for _, input := range job.Task.Inputs {
		err := mapper.MapInput(job.JobID, input.Location, input.Path, input.Class)
		if err != nil {
			return err
		}
	}

	// MapOutput finds where to output the results, and adds that
	// to Job. It also sets that output path should be the output
	// location once the job is done.
	for _, output := range job.Task.Outputs {
		err := mapper.MapOutput(job.JobID, output.Location, output.Path, output.Class, output.Create)
		if err != nil {
			return err
		}
	}

	// TODO allow context to cancel the step loop
	for stepNum, dockerTask := range job.Task.Docker {
		stdin, err := FindStdin(mapper.jobs[job.JobID].Bindings, dockerTask.Stdin)
		if err != nil {
			return fmt.Errorf("Error setting up job stdin: %s", err)
		}

		// Create file for job stdout
		stdout, err := mapper.TempFile(job.JobID)
		if err != nil {
			return fmt.Errorf("Error setting up job stdout log: %s", err)
		}

		// Create file for job stderr
		stderr, err := mapper.TempFile(job.JobID)
		if err != nil {
			return fmt.Errorf("Error setting up job stderr log: %s", err)
		}
		stdoutPath := stdout.Name()
		stderrPath := stderr.Name()

		// `binds` is a slice of the docker run arguments.
		binds := mapper.GetBindings(job.JobID)

		dcmd := DockerCmd{
			ImageName:       dockerTask.ImageName,
			Cmd:             dockerTask.Cmd,
			Binds:           binds,
			Workdir:         dockerTask.Workdir,
			RemoveContainer: true,
			Stdin:           stdin,
			Stdout:          stdout,
			Stderr:          stderr,
		}
		cmdErr := dcmd.Run()
		exitCode := getExitCode(cmdErr)
		log.Printf("Exit code: %d", exitCode)

		stdout.Close()
		stderr.Close()

		// If `Stderr` is supposed to be added to the volume, copy it.
		if dockerTask.Stderr != "" {
			hstPath := mapper.HostPath(job.JobID, dockerTask.Stderr)
			if len(hstPath) > 0 {
				copyFileContents(stderrPath, hstPath)
			}

		}
		//If `Stdout` is supposed to be added to the volume, copy it.
		if dockerTask.Stdout != "" {
			hstPath := mapper.HostPath(job.JobID, dockerTask.Stdout)
			if len(hstPath) > 0 {
				copyFileContents(stdoutPath, hstPath)
			}
		}

		stderrText := readFileHead(stderrPath)
		stdoutText := readFileHead(stdoutPath)

		// Send the scheduler service a job status update
		statusReq := &ga4gh_task_ref.UpdateStatusRequest{
			Id:   job.JobID,
			Step: int64(stepNum),
			Log: &ga4gh_task_exec.JobLog{
				Stdout:   string(stdoutText),
				Stderr:   string(stderrText),
				ExitCode: int32(exitCode),
			},
		}
		eng.sched.UpdateJobStatus(ctx, statusReq)

		if cmdErr != nil {
			return cmdErr
		}
	}

	mapper.FinalizeJob(job.JobID)

	return nil
}

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitCode(err error) int {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return status.ExitStatus()
			}
		} else {
			log.Printf("Could not determine exit code. Using default -999")
			return -999
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
}

// TODO I'm not sure what the best place for this is, or how/when is best to create
//      the file mapper/client.
func getFileClient(config FileConfig) FileSystemAccess {

	if config.VolumeDir != "" {
		volumeDir, _ := filepath.Abs(config.VolumeDir)
		if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
			os.Mkdir(volumeDir, 0700)
		}
	}

	// OpenStack Swift object storage
	if config.SwiftCacheDir != "" {
		// Mock Swift storage directory to local filesystem.
		// NOT actual swift.
		storageDir, _ := filepath.Abs(config.SwiftCacheDir)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}

		return NewSwiftAccess()

		// Local filesystem storage
	} else if config.AllowedDirs != "" {
		o := []string{}
		for _, i := range strings.Split(config.AllowedDirs, ",") {
			p, _ := filepath.Abs(i)
			o = append(o, p)
		}
		return NewFileAccess(o)

		// Shared filesystem storage
	} else if config.SharedDir != "" {
		storageDir, _ := filepath.Abs(config.SharedDir)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}
		return NewSharedFS(storageDir)

	} else {
		// TODO what's a good default? Or error?
		return NewSharedFS("storage")
	}
}
