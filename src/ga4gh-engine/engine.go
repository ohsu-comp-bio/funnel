package ga4gh_taskengine

import (
	"os"
	"github.com/google/shlex"
	"fmt"
	"ga4gh-tasks"
)


const HEADER_SIZE = int64(102400)
func read_file_head(path string) []byte {
	f, _ := os.Open(path)
	buffer := make([]byte, HEADER_SIZE)
	l, _ := f.Read(buffer)
	f.Close()
	return buffer[:l]
}


func RunJob(job *ga4gh_task_exec.TaskOp, mapper FileMapper) error {

	for _, input := range job.Task.InputParameters {
		if job.TaskArgs == nil {
			return fmt.Errorf("No task arguments found")
		}
		srcPath := job.TaskArgs.Inputs[input.Name]
		mapper.MapInput(job.TaskOpId, srcPath, *input.LocalCopy)
	}

	for _, output := range(job.Task.OutputParameters) {
		dstPath := job.TaskArgs.Outputs[output.Name]
		mapper.MapOutput(job.TaskOpId, *output.LocalCopy, dstPath)
	}

	for i, dockerTask := range job.Task.Docker {
		stdout, err := mapper.TempFile(job.TaskOpId)
		if err != nil {
			return fmt.Errorf("Error setting up job stdout log: %s", err)
			return err
		}
		stderr, err := mapper.TempFile(job.TaskOpId)
		if err != nil {
			return fmt.Errorf("Error setting up job stderr log: %s", err)
			return err
		}
		stdout_path := stdout.Name()
		stderr_path := stderr.Name()
		if err != nil {
			return fmt.Errorf("Error setting up job")
		}
		binds := mapper.GetBindings(job.TaskOpId)

		dclient := NewDockerDirect()
		cmds, _ := shlex.Split(dockerTask.Cmd)
		exit_code, err := dclient.Run(dockerTask.ImageName, cmds, binds, true, stdout, stderr)
		stdout.Close()
		stderr.Close()

		stderr_text := read_file_head(stderr_path)
		stdout_text := read_file_head(stdout_path)
		mapper.UpdateOutputs(job.TaskOpId, i, exit_code, string(stdout_text), string(stderr_text))
		if err != nil {
			return err
		}
	}


	mapper.FinalizeJob(job.TaskOpId)

	return nil
}