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
		if exec.Image == "" {
			errs.add("Task.Executors[%d].Image: required, but empty", i)
		}

		if len(exec.Command) == 0 {
			errs.add("Task.Executors[%d].Command: required, but empty", i)
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
	}

	for i, input := range t.Inputs {
		if input.Content != "" && input.Url != "" {
			errs.add("Task.Inputs[%d].Content: Url is non-empty", i)
		} else if input.Url == "" && input.Content == "" {
			errs.add("Task.Inputs[%d].Url: required, but empty", i)
		}

		if input.Path == "" {
			errs.add("Task.Inputs[%d].Path: required, but empty", i)
		}

		if input.Path != "" && !strings.HasPrefix(input.Path, "/") {
			errs.add("task.Inputs[%d].Path: must be an absolute path", i)
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
	}

	for i, vol := range t.Volumes {
		if !strings.HasPrefix(vol, "/") {
			errs.add("Task.Volumes[%d]: must be an absolute path", i)
		}
	}

	for k, v := range t.Tags {
		if k == "" {
			errs.add(`Task.Tags[""]=%s: empty key`, v)
		}
	}

	return errs
}
