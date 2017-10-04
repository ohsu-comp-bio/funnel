package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/manual"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"github.com/ohsu-comp-bio/funnel/worker"
	"testing"
	"time"
)

func TestNodeUpdated(t *testing.T) {
	conf := DefaultConfig()
	conf.Scheduler.Node.ID = "node-1"
	conf.Scheduler.Node.Resources.Cpus = 10
	conf.Scheduler.Node.Resources.RamGb = 100
	conf.Scheduler.Node.Resources.DiskGb = 1000
	conf.Scheduler.Node.UpdateRate = time.Millisecond * 10
	conf.Scheduler.ScheduleRate = time.Millisecond * 5
	fun := NewFunnel(conf)

	man, err := manual.NewBackend(conf)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a dummy worker that sleeps
	w := &worker.NoopWorker{
		OnRun: func(context.Context, *tes.Task) {
			log.Debug("SLEEPING")
			time.Sleep(time.Second)
		},
	}

	n, err := scheduler.NewNode(conf, w)
	if err != nil {
		t.Fatal(err)
	}

	sch := scheduler.NewScheduler(fun.SDB, man, conf.Scheduler)
	sback := scheduler.NewComputeBackend(fun.SDB)
	fun.DB.WithComputeBackend(sback)
	fun.Scheduler = sch
	fun.StartServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go n.Run(ctx)

	time.Sleep(time.Millisecond * 100)

	nodes := fun.ListNodes()
	if len(nodes) != 1 {
		t.Error("unexpected node count", len(nodes))
	}

	n0 := nodes[0]
	n0r := n0.Resources
	n0a := n0.Available

	if n0r.Cpus != 10 || n0r.RamGb != 100 || n0r.DiskGb != 1000 {
		t.Error("unexpected resources", n0r)
	}

	if n0a.Cpus != 10 || n0a.RamGb != 100 || n0a.DiskGb != 1000 {
		t.Error("unexpected available resources", n0a)
	}

	fun.Run(`
    --sh 'echo hello'
    --cpu 5
    --ram 50
    --disk 500
  `)

	time.Sleep(time.Millisecond * 100)

	nodes = fun.ListNodes()
	if len(nodes) != 1 {
		t.Error("unexpected node count", len(nodes))
	}

	n0 = nodes[0]
	n0r = n0.Resources
	n0a = n0.Available

	if n0r.Cpus != 10 || n0r.RamGb != 100 || n0r.DiskGb != 1000 {
		t.Error("unexpected resources", n0r)
	}

	if n0a.Cpus != 5 || n0a.RamGb != 50 || n0a.DiskGb != 500 {
		t.Error("unexpected resources", n0a)
	}
}

// Test a scheduled task is removed from the task queue.
func TestScheduledTaskRemovedFromQueue(t *testing.T) {
	ctx := context.Background()
	conf := DefaultConfig()

	db, err := boltdb.NewBoltDB(conf)
	if err != nil {
		panic(err)
	}

	db.PutNode(ctx, &pbs.Node{
		Id:    "node-1",
		State: pbs.NodeState_ALIVE,
	})

	// Get the node back, so we have the correct db version.
	node, err := db.GetNode(ctx, &pbs.GetNodeRequest{Id: "node-1"})
	if err != nil {
		panic(err)
	}

	// Set up dummy backend that makes an offer.
	backend := &dummyBackend{node}
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
	sch.Schedule(ctx)

	res2 := db.ReadQueue(10)
	if len(res2) != 0 {
		t.Fatal("Expected task queue to be empty")
	}
}

type dummyBackend struct {
	n *pbs.Node
}

func (d *dummyBackend) GetOffer(t *tes.Task) *scheduler.Offer {
	return scheduler.NewOffer(d.n, t, nil)
}
