package e2e

import (
	"context"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/worker"
	"testing"
)

func TestWorkerCmdRun(t *testing.T) {
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()

	// this only writes the task to the DB since the 'noop'
	// compute backend is in use
	id := f.Run(`
    --sh 'echo hello world'
  `)

	err := workerCmd.Run(c.Worker, id)
	if err != nil {
		log.Error("err", err)
		t.Fatal("unexpected error")
	}

	task, err := f.HTTP.GetTask(id, "FULL")
	if err != nil {
		log.Error("err", err)
		t.Fatal("unexpected error")
	}

	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected state")
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing stdout")
	}
}

func TestDefaultWorkerRun(t *testing.T) {
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()

	// this only writes the task to the DB since the 'noop'
	// compute backend is in use
	id := f.Run(`
    --sh 'echo hello world'
  `)

	w, err := worker.NewDefaultWorker(c.Worker, id)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	w.Run(context.Background())
	f.Wait(id)

	task, err := f.HTTP.GetTask(id, "FULL")
	if err != nil {
		log.Error("err", err)
		t.Fatal("unexpected error")
	}

	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected state")
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing stdout")
	}
}
