package tests

import (
	"errors"
	"funnel/logger"
	"funnel/proto/tes"
	"golang.org/x/net/context"
	"testing"
	"time"
)

var log = logger.New("tests")

func init() {
	logger.ForceColors()
}

// Test the flow of a task being scheduled to a worker, run, completed, etc.
func TestBasicWorker(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()
	ctx := context.Background()

	// Run task
	taskID := srv.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[taskID]

	if ctrl == nil {
		t.Error("Expected controller for task")
		return
	}

	if ctrl.State() != tes.State_INITIALIZING {
		t.Error("Expected runner state to be init")
		return
	}

	// Set task to running and sync worker
	ctrl.SetRunning()
	srv.Flush()

	// Check task state in DB
	r, _ := srv.DB.GetTask(ctx, &tes.GetTaskRequest{Id: taskID})

	if r.State != tes.State_RUNNING {
		t.Error("Expected task state in DB to be running")
		return
	}

	// Set task to complete and sync worker
	ctrl.SetResult(nil)
	srv.Flush()

	// Check for complete state in database
	q, _ := srv.DB.GetTask(ctx, &tes.GetTaskRequest{Id: taskID})

	if q.State != tes.State_COMPLETE {
		t.Error("Expected task state in DB to be running")
		return
	}
	log.Debug("TEST", "taskID", taskID, "r", r)
}

// Test a scheduled task is removed from the task queue.
func TestScheduledTaskRemovedFromQueue(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()

	srv.RunHelloWorld()
	srv.Flush()

	res := srv.DB.ReadQueue(10)
	if len(res) != 0 {
		t.Error("Expected task queue to be empty")
		return
	}
}

// Test the case where a task fails.
func TestTaskFail(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()
	ctx := context.Background()

	// Run task
	taskID := srv.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[taskID]

	if ctrl == nil {
		t.Error("Expected controller for task")
		return
	}

	// Set failed and sync
	ctrl.SetResult(errors.New("TEST"))
	srv.Flush()

	// Check task state in DB
	r, _ := srv.DB.GetTask(ctx, &tes.GetTaskRequest{Id: taskID})

	if r.State != tes.State_ERROR {
		t.Error("Expected task state in DB to be running")
		return
	}

	// Sync worker. The worker should remove the task controller for the
	// failed task.
	srv.Flush()
	// There was a bug where the worker was re-running failed tasks.
	// Do a few syncs just to make sure.
	srv.Flush()
	srv.Flush()

	if len(srv.NoopWorker.Ctrls) != 0 {
		t.Error("Expected task control to be cleaned up.")
		return
	}
}

// Test the flow of a worker completing a task then timing out
func TestWorkerTimeout(t *testing.T) {
	conf := NewConfig()
	conf.Worker.Timeout = time.Millisecond
	conf.Worker.UpdateRate = time.Millisecond * 2

	srv := NewFunnel(conf)
	srv.Start()
	defer srv.Stop()

	done := make(chan struct{})
	go func() {
		srv.NoopWorker.Run()
		log.Debug("DONE")
		close(done)
	}()

	taskID := srv.RunHelloWorld()

	// Sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[taskID]

	if ctrl == nil {
		t.Error("Expected controller for task")
		return
	}

	// Set task complete
	ctrl.SetResult(nil)
	srv.Flush()
	srv.Flush()

	timeout := time.NewTimer(conf.Worker.Timeout * 100)

	// Wait for either the worker to be done, or the test to timeout
	select {
	case <-timeout.C:
		t.Error("Expected worker to be done")
	case <-done:
		// Worker is done
	}
}
