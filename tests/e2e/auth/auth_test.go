package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"strings"
	"testing"
)

var extask = &tes.Task{
	Executors: []*tes.Executor{
		{
			ImageName: "alpine",
			Cmd:       []string{"echo", "hello world"},
		},
	},
}

func TestBasicAuthFail(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Server.Password = "abc123"
	fun := e2e.NewFunnel(conf)
	fun.StartServer()

	var err error
	_, err = fun.HTTP.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   "1",
		View: tes.TaskView_MINIMAL,
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.ListTasks(context.Background(), &tes.ListTasksRequest{
		View: tes.TaskView_MINIMAL,
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CreateTask(context.Background(), extask)
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CancelTask(context.Background(), &tes.CancelTaskRequest{
		Id: "1",
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.RunE(`--sh 'echo hello'`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBasicAuthed(t *testing.T) {
	os.Setenv("FUNNEL_SERVER_PASSWORD", "abc123")
	defer os.Unsetenv("FUNNEL_SERVER_PASSWORD")

	conf := e2e.DefaultConfig()
	conf.Server.Password = "abc123"
	fun := e2e.NewFunnel(conf)
	fun.StartServer()
	fun.AddRPCClient(util.PerRPCPassword("abc123"))

	var err error

	// Run a task to completion
	id2 := fun.Run(`--sh 'echo hello'`)
	t2 := fun.Wait(id2)
	if t2.State != tes.State_COMPLETE {
		t.Fatal("expected task to complete")
	}

	_, err = fun.HTTP.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id2,
		View: tes.TaskView_MINIMAL,
	})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	_, err = fun.HTTP.ListTasks(context.Background(), &tes.ListTasksRequest{
		View: tes.TaskView_MINIMAL,
	})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	resp, err := fun.HTTP.CreateTask(context.Background(), extask)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	_, err = fun.HTTP.CancelTask(context.Background(), &tes.CancelTaskRequest{
		Id: resp.Id,
	})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}
