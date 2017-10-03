package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"testing"
)

// Test a scheduled task is removed from the task queue.
func TestScheduledTaskRemovedFromQueue(t *testing.T) {
	ctx := context.Background()
	conf := DefaultConfig()

	db, err := boltdb.NewBoltDB(conf)
	if err != nil {
		panic(err)
	}

	// Set up dummy backend that makes an offer.
	backend := dummyBackend{
		offerfunc: func(t *tes.Task) *scheduler.Offer {
			return scheduler.NewOffer(&pbs.Node{Id: "node-1"}, t, nil)
		},
	}

	sch := scheduler.NewScheduler(db, backend, conf.Scheduler)
	sback := scheduler.NewComputeBackend(db)
	db.WithComputeBackend(sback)

	task := &tes.Task{
		Id: "task-1",
		Executors: []*tes.Executor{
			{
				ImageName: "alpine",
				Cmd:       []string{"hello"},
			},
		},
	}

	_, err = db.CreateTask(ctx, task)
	if err != nil {
		t.Fatal("CreateTask failed", err)
	}

	res := db.ReadQueue(10)
	if len(res) != 1 {
		t.Fatal("Expected task in queue")
	}

	db.PutNode(ctx, &pbs.Node{
		Id:    "node-1",
		State: pbs.NodeState_ALIVE,
	})
	sch.Schedule(ctx)

	res2 := db.ReadQueue(10)
	if len(res2) != 0 {
		t.Fatal("Expected task queue to be empty")
	}
}
