package tesTaskEngineWorker

import (
	"fmt"
	"os"
	"tes/ga4gh"
)

const headerSize = int64(102400)

func readFileHead(path string) []byte {
	f, _ := os.Open(path)
	buffer := make([]byte, headerSize)
	l, _ := f.Read(buffer)
	f.Close()
	return buffer[:l]
}

// RunJob runs a job.
func RunJob(job *ga4gh_task_exec.Job, mapper FileMapper) error {
	// Modifies the filemapper's jobID
	mapper.Job(job.JobId)

	// Iterates through job.Task.Resources.Volumes and add the volume to mapper.
	for _, disk := range job.Task.Resources.Volumes {
		mapper.AddVolume(job.JobId, disk.Source, disk.MountPoint)
	}

	// MapInput copies the input.Location into input.Path.
	for _, input := range job.Task.Inputs {
		err := mapper.MapInput(job.JobId, input.Location, input.Path, input.Class)
		if err != nil {
			return err
		}
	}

	// MapOutput finds where to output the results, and adds that
	// to Job. It also sets that output path should be the output
	// location once the job is done.
	for _, output := range job.Task.Outputs {
		err := mapper.MapOutput(job.JobId, output.Location, output.Path, output.Class, output.Create)
		if err != nil {
			return err
		}
	}

	// Loops through Docker Tasks, and label them with i (the index).
	for i, dockerTask := range job.Task.Docker {
		// Finds stdout path through mapper.TempFile.
		// Takes stdout from Tool, and outputs into a file.
		stdout, err := mapper.TempFile(job.JobId)
		if err != nil {
			return fmt.Errorf("Error setting up job stdout log: %s", err)
		}
		// Finds stderr path through mapper.TempFile.
		// Takes stderr from Tool, and outputs into a file.
		// `stderr` is a stream where systems error is saved.
		// `err` is Go error.
		stderr, err := mapper.TempFile(job.JobId)
		if err != nil {
			// why two returns? will second return actually return?
			return fmt.Errorf("Error setting up job stderr log: %s", err)
		}
		stdoutPath := stdout.Name()
		stderrPath := stderr.Name()

		// `binds` is a slice of the docker run arguments.
		binds := mapper.GetBindings(job.JobId)

		// NewDockerEngine returns a type that has a `Run` method.
		dclient := NewDockerEngine()

		// ImageName == Docker image name (ex. devian:Wheezy).
		// cmd = Docker command (ex. `cat`).
		// workdir = Docker working directory (ex. /mnt/work).
		exitCode, err := dclient.Run(dockerTask.ImageName, dockerTask.Cmd, binds, dockerTask.Workdir, true, stdout, stderr)

		stdout.Close()
		stderr.Close()

		// If `Stderr` is supposed to be added to the volume, copy it.
		if len(dockerTask.Stderr) > 0 {
			hstPath := mapper.HostPath(job.JobId, dockerTask.Stderr)
			if len(hstPath) > 0 {
				copyFileContents(stderrPath, hstPath)
			}

		}
		//If `Stdout` is supposed to be added to the volume, copy it.
		if len(dockerTask.Stdout) > 0 {
			hstPath := mapper.HostPath(job.JobId, dockerTask.Stdout)
			if len(hstPath) > 0 {
				copyFileContents(stdoutPath, hstPath)
			}
		}

		stderrText := readFileHead(stderrPath)
		stdoutText := readFileHead(stdoutPath)
		mapper.UpdateOutputs(job.JobId, i, exitCode, string(stdoutText), string(stderrText))
		if err != nil {
			return err
		}
	}

	mapper.FinalizeJob(job.JobId)

	return nil
}
