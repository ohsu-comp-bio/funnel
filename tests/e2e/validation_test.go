package e2e

import (
	"github.com/andreyvit/diff"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"google.golang.org/grpc"
	"strings"
	"testing"
)

func TestTaskNoExecutorsValidationError(t *testing.T) {
	setLogOutput(t)
	_, err := fun.RunTask(&tes.Task{})
	if err == nil {
		t.Error("expected validation error")
	}
	expected := `invalid task message:
Task.Executors: at least one executor is required
`

	e := grpc.ErrorDesc(err)
	if e != expected {
		t.Error("unexpected error message", diff.LineDiff(e, expected))
	}
}

func TestTaskInputContentsValidationError(t *testing.T) {
	setLogOutput(t)
	_, err := fun.RunTask(&tes.Task{
		Inputs: []*tes.TaskParameter{
			{
				Contents: "foo",
				Url:      "bar",
			},
		},
	})
	if err == nil {
		t.Error("expected validation error")
	}

	e := grpc.ErrorDesc(err)

	if !strings.Contains(e, "Task.Inputs[0].Contents") {
		t.Error("unexpected error message")
	}
}

func TestTaskInputContentsValidation(t *testing.T) {
	setLogOutput(t)
	_, err := fun.RunTask(&tes.Task{
		Executors: []*tes.Executor{
			{
				ImageName: "alpine",
				Cmd:       []string{"echo"},
			},
		},
		Inputs: []*tes.TaskParameter{
			{
				Contents: "foo",
				Url:      "",
				Path:     "/bar",
			},
		},
	})
	if err != nil {
		t.Error("unexpected validation error")
	}
}

func TestTaskValidationError(t *testing.T) {
	setLogOutput(t)
	_, err := fun.RunTask(&tes.Task{
		Executors: []*tes.Executor{
			{},
		},
		Inputs: []*tes.TaskParameter{
			{},
		},
		Outputs: []*tes.TaskParameter{
			{},
		},
		Volumes: []string{"not-absolute"},
	})

	if err == nil {
		t.Error("expected validation error")
	}

	expected := `invalid task message:
Task.Executors[0].ImageName: required, but empty
Task.Executors[0].Cmd: required, but empty
Task.Inputs[0].Url: required, but empty
Task.Inputs[0].Path: required, but empty
Task.Outputs[0].Url: required, but empty
Task.Outputs[0].Path: required, but empty
Task.Volumes[0]: must be an absolute path
`

	e := grpc.ErrorDesc(err)
	if e != expected {
		t.Errorf("expected volumes error message. diff:\n%s", diff.LineDiff(e, expected))
	}
}
