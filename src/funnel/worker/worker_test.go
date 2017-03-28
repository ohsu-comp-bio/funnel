package worker

import (
	"errors"
	"funnel/config"
	pbe "funnel/ga4gh"
	"funnel/logger"
	sched_mocks "funnel/scheduler/mocks"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func init() {
	logger.ForceColors()
}

// Test the flow of a job being scheduled to a worker, run, completed, etc.
func TestBasicWorker(t *testing.T) {
	// Set up
	srv := newMockSchedulerServer()
	defer srv.Close()
	ctx := context.Background()

	// Run task
	jobID := srv.Server.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.worker.Ctrls[jobID]

	if ctrl == nil {
		t.Error("Expected controller for job")
	}

	if ctrl.State() != pbe.State_Initializing {
		t.Error("Expected runner state to be init")
	}

	// Set job to running and sync worker
	ctrl.SetRunning()
	srv.Flush()

	// Check job state in DB
	r, _ := srv.db.GetJob(ctx, &pbe.JobID{Value: jobID})

	if r.State != pbe.State_Running {
		t.Error("Expected job state in DB to be running")
	}

	// Set job to complete and sync worker
	ctrl.SetResult(nil)
	srv.Flush()

	// Check for complete state in database
	q, _ := srv.db.GetJob(ctx, &pbe.JobID{Value: jobID})

	if q.State != pbe.State_Complete {
		t.Error("Expected job state in DB to be running")
	}
	log.Debug("TEST", "jobID", jobID, "r", r)
}

// Test a scheduled job is removed from the job queue.
// TODO doesn't this belong more in the scheduler?
func TestScheduledJobRemovedFromQueue(t *testing.T) {
	srv := newMockSchedulerServer()
	defer srv.Close()

	srv.Server.RunHelloWorld()
	srv.Flush()

	res := srv.db.ReadQueue(10)
	if len(res) != 0 {
		t.Error("Expected job queue to be empty")
	}
}

// Test the case where a job fails.
func TestJobFail(t *testing.T) {
	// Set up
	srv := newMockSchedulerServer()
	defer srv.Close()
	ctx := context.Background()

	// Run task
	jobID := srv.Server.RunHelloWorld()

	// Schedule and sync worker
	srv.Flush()
	ctrl := srv.worker.Ctrls[jobID]

	// Set failed and sync
	ctrl.SetResult(errors.New("TEST"))
	srv.Flush()

	// Check job state in DB
	r, _ := srv.db.GetJob(ctx, &pbe.JobID{Value: jobID})

	if r.State != pbe.State_Error {
		t.Error("Expected job state in DB to be running")
	}

	// Sync worker. The worker should remove the job controller for the
	// failed job.
	srv.Flush()
	// There was a bug where the worker was re-running failed jobs.
	// Do a few syncs just to make sure.
	srv.Flush()
	srv.Flush()

	if len(srv.worker.Ctrls) != 0 {
		t.Error("Expected job control to be cleaned up.")
	}
}

// Mainly exercising a panic bug caused by an unhandled
// error from client.GetWorker().
func TestGetWorkerFail(t *testing.T) {
	// Create worker
	conf := config.WorkerDefaultConfig(config.DefaultConfig())
	w, err := NewWorker(conf)
	if err != nil {
		t.Error(err)
	}

	// Override worker client with new mock
	m := new(sched_mocks.Client)
	s := &schedClient{m, conf}
	w.sched = s

	// Set GetWorker to return an error
	m.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("TEST"))
	w.Sync()
}

// Test the flow of a worker completing a job then timing out
func TestWorkerTimeout(t *testing.T) {
	conf := config.DefaultConfig()
	conf.Worker.Timeout = time.Millisecond
	conf.Worker.UpdateRate = time.Millisecond * 2
	srv := MockSchedulerServerFromConfig(conf)
	defer srv.Close()

	done := make(chan struct{})
	go func() {
		srv.worker.Run()
		log.Debug("DONE")
		close(done)
	}()

	jobID := srv.Server.RunHelloWorld()

	// Sync worker
	srv.Flush()
	ctrl := srv.worker.Ctrls[jobID]

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

// Test calling Worker.Stop()
func TestStopWorker(t *testing.T) {
	w, _ := NewWorker(config.WorkerDefaultConfig(config.DefaultConfig()))
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	timeout := time.NewTimer(time.Millisecond * 4)
	w.Stop()

	// Wait for either the worker to be done, or the test to timeout
	select {
	case <-timeout.C:
		t.Error("Expected worker to be done")
	case <-done:
		// Worker is done
	}
}
