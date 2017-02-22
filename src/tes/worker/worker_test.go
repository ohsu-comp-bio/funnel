package worker

import (
	"golang.org/x/net/context"
	pbe "tes/ga4gh"
	"testing"
	//"tes/config"
	//pbr "tes/server/proto"
	"tes/logger"
)

func init() {
	logger.ForceColors()
}

func TestBasicWorker(t *testing.T) {
	srv := newMockSchedulerServer()
	defer srv.Stop()
	ctx := context.Background()
	var err error

	jid, err := srv.db.RunTask(ctx, &pbe.Task{
		Name: "test-task",
		Docker: []*pbe.DockerExecutor{
			{},
		},
	})
	if err != nil {
		panic(err)
	}
	jobID := jid.Value

	srv.Flush()
	ctrl := srv.worker.ctrls[jobID]

	if ctrl == nil {
		t.Error("Expected controller for job")
	}

	if ctrl.State() != pbe.State_Initializing {
		t.Error("Expected runner state to be init")
	}

	ctrl.SetRunning()
	srv.Flush()

	r, err := srv.db.GetJob(ctx, &pbe.JobID{Value: jobID})

	if r.State != pbe.State_Running {
		t.Error("Expected job state in DB to be running")
	}

	ctrl.SetResult(nil)
	srv.Flush()

	q, err := srv.db.GetJob(ctx, &pbe.JobID{Value: jobID})

	if q.State != pbe.State_Complete {
		t.Error("Expected job state in DB to be running")
	}
	log.Debug("TEST", "jobID", jobID, "r", r)
}

func TestScheduledJobRemovedFromQueue(t *testing.T) {
	srv := newMockSchedulerServer()
	defer srv.Stop()
	ctx := context.Background()
	var err error

	_, err = srv.db.RunTask(ctx, &pbe.Task{
		Name: "test-task",
		Docker: []*pbe.DockerExecutor{
			{},
		},
	})
	if err != nil {
		panic(err)
	}

	srv.Flush()

	res := srv.db.ReadQueue(10)
	if len(res) != 0 {
		t.Error("Expected job queue to be empty")
	}
}
