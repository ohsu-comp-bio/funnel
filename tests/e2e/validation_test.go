package e2e

import (
	"github.com/andreyvit/diff"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"google.golang.org/grpc"
	"strings"
	"testing"
)

func TestTaskNoExecutorsValidationError(t *testing.T) {
	_, err := fun.RunTask(&tes.Task{})
	log.Debug("VALID ERR", err)
	if err == nil {
		t.Fatal("expected validation error")
	}
	expected := `Task.Executors: at least one executor is required
`

	e := grpc.ErrorDesc(err)
	log.Debug("VALID ERR", e)
	if e != expected {
		log.Debug("DIFF", diff.LineDiff(e, expected))
		t.Fatal("unexpected error message")
	}
}

func TestTaskInputContentsValidationError(t *testing.T) {
	_, err := fun.RunTask(&tes.Task{
		Inputs: []*tes.TaskParameter{
			{
				Contents: "foo",
				Url:      "bar",
			},
		},
	})
	log.Debug("VALID ERR", err)
	if err == nil {
		t.Fatal("expected validation error")
	}

	e := grpc.ErrorDesc(err)
	log.Debug("VALID ERR", e)

	if !strings.Contains(e, "Task.Inputs[0].Contents") {
		t.Fatal("unexpected error message")
	}
}

func TestTaskInputContentsValidation(t *testing.T) {
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
		t.Fatal("unexpected validation error")
	}
}

func TestTaskValidationError(t *testing.T) {
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
		t.Fatal("expected validation error")
	}

	expected := `Task.Executors[0].ImageName: required, but empty
Task.Executors[0].Cmd: required, but empty
Task.Inputs[0].Url: required, but empty
Task.Inputs[0].Path: required, but empty
Task.Outputs[0].Url: required, but empty
Task.Outputs[0].Path: required, but empty
Task.Volumes[0]: must be an absolute path
`

	e := grpc.ErrorDesc(err)
	log.Debug("VALID ERR", e)
	if e != expected {
		log.Debug("DIFF", diff.LineDiff(e, expected))
		t.Fatal("expected volumes error message")
	}
}
