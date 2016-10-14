package ga4gh_taskengine_worker

import (
	"fmt"
	"tes/ga4gh"
	"os"
)

const HEADER_SIZE = int64(102400)

func read_file_head(path string) []byte {
	f, _ := os.Open(path)
	buffer := make([]byte, HEADER_SIZE)
	l, _ := f.Read(buffer)
	f.Close()
	return buffer[:l]
}

func RunJob(job *ga4gh_task_exec.Job, mapper FileMapper) error {

	mapper.Job(job.JobId)

	for _, disk := range job.Task.Resources.Volumes {
		mapper.AddVolume(job.JobId, disk.Source, disk.MountPoint)
	}

	for _, input := range job.Task.Inputs {
		err := mapper.MapInput(job.JobId, input.Location, input.Path, input.Class)
		if err != nil {
			return err
		}
	}

	for _, output := range job.Task.Outputs {
		err := mapper.MapOutput(job.JobId, output.Location, output.Path, output.Class, output.Create)
		if err != nil {
			return err
		}
	}

	for i, dockerTask := range job.Task.Docker {
		stdout, err := mapper.TempFile(job.JobId)
		if err != nil {
			return fmt.Errorf("Error setting up job stdout log: %s", err)
			return err
		}
		stderr, err := mapper.TempFile(job.JobId)
		if err != nil {
			return fmt.Errorf("Error setting up job stderr log: %s", err)
			return err
		}
		stdout_path := stdout.Name()
		stderr_path := stderr.Name()
		if err != nil {
			return fmt.Errorf("Error setting up job")
		}
		binds := mapper.GetBindings(job.JobId)

		dclient := NewDockerEngine()
		exit_code, err := dclient.Run(dockerTask.ImageName, dockerTask.Cmd, binds, dockerTask.Workdir, true, stdout, stderr)
		stdout.Close()
		stderr.Close()

		//If the STDERR is supposed to be added to the volume, copy it in
		if len(dockerTask.Stderr) > 0 {
			hstPath := mapper.HostPath(job.JobId, dockerTask.Stderr)
			if len(hstPath) > 0 {
				copyFileContents(stderr_path, hstPath)
			}

		}
		//If the STDERR is supposed to be added to the volume, copy it in
		if len(dockerTask.Stdout) > 0 {
			hstPath := mapper.HostPath(job.JobId, dockerTask.Stdout)
			if len(hstPath) > 0 {
				copyFileContents(stdout_path, hstPath)
			}
		}

		stderr_text := read_file_head(stderr_path)
		stdout_text := read_file_head(stdout_path)
		mapper.UpdateOutputs(job.JobId, i, exit_code, string(stdout_text), string(stderr_text))
		if err != nil {
			return err
		}
	}

	mapper.FinalizeJob(job.JobId)

	return nil
}
