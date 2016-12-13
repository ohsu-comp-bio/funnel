package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"tes/ga4gh"
	"tes/server/proto"
)

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

// RunJob runs a job.
func RunJob(sched ga4gh_task_ref.SchedulerClient, job *ga4gh_task_exec.Job, mapper FileMapper) error {
	// Modifies the filemapper's jobID
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
		// TODO context should be created at the top-level and passed down
		ctx := context.Background()
		sched.UpdateJobStatus(ctx, statusReq)

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
