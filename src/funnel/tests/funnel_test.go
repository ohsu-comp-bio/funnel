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

// Test the flow of a job being scheduled to a worker, run, completed, etc.
func TestBasicWorker(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()
	ctx := context.Background()

	// Run task
	jobID := srv.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[jobID]

	if ctrl == nil {
		t.Error("Expected controller for job")
		return
	}

	if ctrl.State() != tes.State_Initializing {
		t.Error("Expected runner state to be init")
		return
	}

	// Set job to running and sync worker
	ctrl.SetRunning()
	srv.Flush()

	// Check job state in DB
	r, _ := srv.DB.GetJob(ctx, &tes.JobID{Value: jobID})

	if r.State != tes.State_Running {
		t.Error("Expected job state in DB to be running")
		return
	}

	// Set job to complete and sync worker
	ctrl.SetResult(nil)
	srv.Flush()

	// Check for complete state in database
	q, _ := srv.DB.GetJob(ctx, &tes.JobID{Value: jobID})

	if q.State != tes.State_Complete {
		t.Error("Expected job state in DB to be running")
		return
	}
	log.Debug("TEST", "jobID", jobID, "r", r)
}

// Test a scheduled job is removed from the job queue.
func TestScheduledJobRemovedFromQueue(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()

	srv.RunHelloWorld()
	srv.Flush()

	res := srv.DB.ReadQueue(10)
	if len(res) != 0 {
		t.Error("Expected job queue to be empty")
		return
	}
}

// Test the case where a job fails.
func TestJobFail(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()
	ctx := context.Background()

	// Run task
	jobID := srv.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[jobID]

	if ctrl == nil {
		t.Error("Expected controller for job")
		return
	}

	// Set failed and sync
	ctrl.SetResult(errors.New("TEST"))
	srv.Flush()

	// Check job state in DB
	r, _ := srv.DB.GetJob(ctx, &tes.JobID{Value: jobID})

	if r.State != tes.State_Error {
		t.Error("Expected job state in DB to be running")
		return
	}

	// Sync worker. The worker should remove the job controller for the
	// failed job.
	srv.Flush()
	// There was a bug where the worker was re-running failed jobs.
	// Do a few syncs just to make sure.
	srv.Flush()
	srv.Flush()

	if len(srv.NoopWorker.Ctrls) != 0 {
		t.Error("Expected job control to be cleaned up.")
		return
	}
}

// Test the flow of a worker completing a job then timing out
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

	jobID := srv.RunHelloWorld()

	// Sync worker
	srv.Flush()
	ctrl := srv.NoopWorker.Ctrls[jobID]

	if ctrl == nil {
		t.Error("Expected controller for job")
		return
	}

	// Set job complete
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
