package tesTaskEngineWorker

import (
	"fmt"
	"os"
	"tes/ga4gh"
)

// HeaderSize documentation
// TODO: documentation
const HeaderSize = int64(102400)

func readFileHead(path string) []byte {
	f, _ := os.Open(path)
	buffer := make([]byte, HeaderSize)
	l, _ := f.Read(buffer)
	f.Close()
	return buffer[:l]
}

// RunJob documentation
// TODO: documentation
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
		}
		stderr, err := mapper.TempFile(job.JobId)
		if err != nil {
			return fmt.Errorf("Error setting up job stderr log: %s", err)
		}
		stdoutPath := stdout.Name()
		stderrPath := stderr.Name()
		if err != nil {
			return fmt.Errorf("Error setting up job")
		}
		binds := mapper.GetBindings(job.JobId)

		dclient := NewDockerEngine()
		exitCode, err := dclient.Run(dockerTask.ImageName, dockerTask.Cmd, binds, dockerTask.Workdir, true, stdout, stderr)
		stdout.Close()
		stderr.Close()

		//If the STDERR is supposed to be added to the volume, copy it in
		if len(dockerTask.Stderr) > 0 {
			hstPath := mapper.HostPath(job.JobId, dockerTask.Stderr)
			if len(hstPath) > 0 {
				copyFileContents(stderrPath, hstPath)
			}

		}
		//If the STDERR is supposed to be added to the volume, copy it in
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
