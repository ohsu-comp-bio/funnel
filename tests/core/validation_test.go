package core

import (
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
	"google.golang.org/grpc/status"
)

func TestTaskNoExecutorsValidationError(t *testing.T) {
	tests.SetLogOutput(log, t)
	_, err := fun.RunTask(&tes.Task{})
	if err == nil {
		t.Error("expected validation error")
	}
	expected := `invalid task message:
Task.Executors: at least one executor is required
`

	e := status.Convert(err).Message()
	if e != expected {
		t.Error("unexpected error message", diff.LineDiff(e, expected))
	}
}

func TestTaskInputContentValidationError(t *testing.T) {
	tests.SetLogOutput(log, t)
	_, err := fun.RunTask(&tes.Task{
		Inputs: []*tes.Input{
			{
				Content: "foo",
				Url:     "bar",
			},
		},
	})
	if err == nil {
		t.Error("expected validation error")
	}

	e := status.Convert(err).Message()

	if !strings.Contains(e, "Task.Inputs[0].Content") {
		t.Error("unexpected error message")
	}
}

func TestTaskInputContentValidation(t *testing.T) {
	tests.SetLogOutput(log, t)
	_, err := fun.RunTask(&tes.Task{
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo"},
			},
		},
		Inputs: []*tes.Input{
			{
				Content: "foo",
				Url:     "",
				Path:    "/bar",
			},
		},
	})
	if err != nil {
		t.Error("unexpected validation error")
	}
}

func TestTaskValidationError(t *testing.T) {
	tests.SetLogOutput(log, t)
	_, err := fun.RunTask(&tes.Task{
		Executors: []*tes.Executor{
			{},
		},
		Inputs: []*tes.Input{
			{},
		},
		Outputs: []*tes.Output{
			{},
		},
		Volumes: []string{"not-absolute"},
	})

	if err == nil {
		t.Error("expected validation error")
	}

	expected := `invalid task message:
Task.Executors[0].Image: required, but empty
Task.Executors[0].Command: required, but empty
Task.Inputs[0].Url: required, but empty
Task.Inputs[0].Path: required, but empty
Task.Outputs[0].Url: required, but empty
Task.Outputs[0].Path: required, but empty
Task.Volumes[0]: must be an absolute path
`

	e := status.Convert(err).Message()
	if e != expected {
		t.Errorf("expected volumes error message. diff:\n%s", diff.LineDiff(e, expected))
	}
}
