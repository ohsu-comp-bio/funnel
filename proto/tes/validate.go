package tes

import (
	"fmt"
	"strings"
)

// ValidationError contains task validation errors.
type ValidationError []error

func (v *ValidationError) add(s string, a ...interface{}) {
	*v = append(*v, fmt.Errorf(s, a...))
}
func (v ValidationError) Error() string {
	s := ""
	for _, e := range v {
		s += fmt.Sprintln(e.Error())
	}
	return s
}

// Validate validates the given task and returns ValidationError,
// or nil if the task is valid.
func Validate(t *Task) ValidationError {
	var errs ValidationError

	if len(t.Executors) == 0 {
		errs.add("Task.Executors: at least one executor is required")
	}

	for i, exec := range t.Executors {
		if exec.ImageName == "" {
			errs.add("Task.Executors[%d].ImageName: required, but empty", i)
		}

		if len(exec.Cmd) == 0 {
			errs.add("Task.Executors[%d].Cmd: required, but empty", i)
		}

		if exec.Workdir != "" && !strings.HasPrefix(exec.Workdir, "/") {
			errs.add("Task.Executors[%d].Workdir: must be an absolute path", i)
		}

		if exec.Stdin != "" && !strings.HasPrefix(exec.Stdin, "/") {
			errs.add("Task.Executors[%d].Stdin: must be an absolute path", i)
		}

		if exec.Stdout != "" && !strings.HasPrefix(exec.Stdout, "/") {
			errs.add("Task.Executors[%d].Stdout: must be an absolute path", i)
		}

		if exec.Stderr != "" && !strings.HasPrefix(exec.Stderr, "/") {
			errs.add("Task.Executors[%d].Stderr: must be an absolute path", i)
		}

		for j, port := range exec.Ports {
			if port.Container == 0 {
				errs.add("Task.Executors[%d].Ports[%d].Container: required, but empty", i, j)
			}
			// TODO spec lists 1024 as minimum. Should that be removed? What if you want port 80?
		}
	}

	for i, input := range t.Inputs {
		if input.Path == "" {
			errs.add("Task.Inputs[%d].Path: required, but empty", i)
		}

		if input.Path != "" && !strings.HasPrefix(input.Path, "/") {
			errs.add("task.Inputs[%d].Path: must be an absolute path", i)
		}

		if input.Contents != "" && input.Url != "" {
			errs.add("Task.Inputs[%d].Contents: Url is non-empty", i)
		} else if input.Url == "" && input.Contents == "" {
			errs.add("Task.Inputs[%d].Url: required, but empty", i)
		}
	}

	for i, output := range t.Outputs {
		if output.Url == "" {
			errs.add("Task.Outputs[%d].Url: required, but empty", i)
		}

		if output.Path == "" {
			errs.add("Task.Outputs[%d].Path: required, but empty", i)
		}

		if output.Path != "" && !strings.HasPrefix(output.Path, "/") {
			errs.add("task.Outputs[%d].Path: must be an absolute path", i)
		}

		if output.Contents != "" {
			errs.add("Task.Outputs[%d].Contents: not allowed", i)
		}
	}

	for i, vol := range t.Volumes {
		if !strings.HasPrefix(vol, "/") {
			errs.add("Task.Volumes[%d]: must be an absolute path", i)
		}
	}

	return errs
}
