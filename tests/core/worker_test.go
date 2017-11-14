package core

import (
	"context"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests"
	"github.com/ohsu-comp-bio/funnel/worker"
	"path"
	"testing"
	"time"
)

func TestWorkerCmdRun(t *testing.T) {
	tests.SetLogOutput(log, t)
	c := tests.DefaultConfig()
	c.Backend = "noop"
	f := tests.NewFunnel(c)
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
	tests.SetLogOutput(log, t)
	c := tests.DefaultConfig()
	c.Backend = "noop"
	f := tests.NewFunnel(c)
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

type eventCounter struct {
	stdout, stderr int
}

func (e *eventCounter) Write(ev *events.Event) error {
	switch ev.Type {
	case events.Type_EXECUTOR_STDOUT:
		e.stdout++
	case events.Type_EXECUTOR_STDERR:
		e.stderr++
	}
	return nil
}
func (e *eventCounter) Close() error {
	return nil
}

type taskReader struct {
	task *tes.Task
}

func (r taskReader) Task() (*tes.Task, error) {
	return r.task, nil
}
func (r taskReader) State() (tes.State, error) {
	return r.task.State, nil
}

// Test that stdout generates events at an expected, consistent rate.
// The task dumps megabytes of random data to stdout. The test waits
// 10 seconds and checks how many stdout events were generated.
func TestLargeLogRate(t *testing.T) {
	tests.SetLogOutput(log, t)
	// Generate 1MB 1000 times to stdout.
	// At the end, echo "\n\nhello\n".
	conf := tests.DefaultConfig().Worker
	conf.UpdateRate = time.Millisecond * 500
	conf.BufferSize = 100
	task := tes.Task{
		Id: "test-task-" + tes.GenerateID(),
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"dd", "if=/dev/urandom", "bs=10000", "count=1"},
			},
		},
	}

	baseDir := path.Join(conf.WorkDir, task.Id)
	reader := taskReader{&task}

	counts := &eventCounter{}
	logger := &events.Logger{Log: log}
	m := events.MultiWriter(logger, counts)

	w := worker.DefaultWorker{
		Conf:       conf,
		Mapper:     worker.NewFileMapper(baseDir),
		Store:      storage.Storage{},
		TaskReader: reader,
		Event:      events.NewTaskWriter(task.Id, 0, conf.Logger.Level, m),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	w.Run(ctx)

	// Given the difficulty of timing how long it task a task + docker container to start,
	// we just check that a small amount of events were generated.
	// 20 events is not too bad for dumping many megabytes of data.
	if counts.stdout > 20 {
		t.Error("unexpected stdout event count", counts.stdout)
	}
}
