package e2e

import (
	"context"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"testing"
)

func TestWorkerCmdRun(t *testing.T) {
	setLogOutput(t)
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()

	// this only writes the task to the DB since the 'noop'
	// compute backend is in use
	id := f.Run(`
    --sh 'echo hello world'
  `)

	err := workerCmd.Run(c.Worker, id, log)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	task, err := f.HTTP.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_FULL,
	})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected state")
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing stdout")
	}
}

func TestDefaultWorkerRun(t *testing.T) {
	setLogOutput(t)
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()

	// this only writes the task to the DB since the 'noop'
	// compute backend is in use
	id := f.Run(`
    --sh 'echo hello world'
  `)

	w, err := workerCmd.NewDefaultWorker(c.Worker, id, log)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	w.Run(context.Background())
	f.Wait(id)

	task, err := f.HTTP.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_FULL,
	})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected state")
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing stdout")
	}
}
