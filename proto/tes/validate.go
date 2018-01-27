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

  if t.Image == "" {
    errs.add("Task.Image: required, but empty")
  }

  if len(t.Command) == 0 {
    errs.add("Task.Command: required, but empty")
  }

  if t.Workdir != "" && !strings.HasPrefix(t.Workdir, "/") {
    errs.add("Task.Workdir: must be an absolute path")
  }

  if t.Stdin != "" && !strings.HasPrefix(t.Stdin, "/") {
    errs.add("Task.Stdin: must be an absolute path")
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

	return errs
}
