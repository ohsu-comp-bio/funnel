package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
	"github.com/ohsu-comp-bio/funnel/util/rpc"
)

var extask = &tes.Task{
	Executors: []*tes.Executor{
		{
			Image:   "alpine",
			Command: []string{"echo", "hello world"},
		},
	},
}

func TestBasicAuthFail(t *testing.T) {
	ctx := context.Background()
	conf := tests.DefaultConfig()
	conf.Server.User = "funnel"
	conf.Server.Password = "abc123"
	fun := tests.NewFunnel(conf)
	fun.StartServer()

	unauthConf := conf.Server
	unauthConf.Password = ""
	conn, err := rpc.Dial(ctx, unauthConf)
	if err != nil {
		t.Fatal(err)
	}
	cli := tes.NewTaskServiceClient(conn)
	defer conn.Close()

	_, err = fun.HTTP.GetTask(ctx, &tes.GetTaskRequest{
		Id:   "1",
		View: tes.TaskView_MINIMAL,
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.ListTasks(ctx, &tes.ListTasksRequest{
		View: tes.TaskView_MINIMAL,
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CreateTask(ctx, extask)
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CancelTask(ctx, &tes.CancelTaskRequest{
		Id: "1",
	})
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		t.Fatal("expected error")
	}

	_, err = cli.CreateTask(ctx, tests.HelloWorld())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBasicAuthed(t *testing.T) {
	os.Setenv("FUNNEL_SERVER_USER", "funnel")
	os.Setenv("FUNNEL_SERVER_PASSWORD", "abc123")
	defer os.Unsetenv("FUNNEL_SERVER_USER")
	defer os.Unsetenv("FUNNEL_SERVER_PASSWORD")

	conf := tests.DefaultConfig()
	conf.Server.User = "funnel"
	conf.Server.Password = "abc123"
	fun := tests.NewFunnel(conf)
	fun.StartServer()

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
