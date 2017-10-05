package node

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
	"time"
)

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead
func TestNodeHealthDead(t *testing.T) {
	ctx := context.Background()
	db, sch := setupSchedTest()

	db.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
	})

	time.Sleep(time.Millisecond * 2)
	err := sch.CheckNodes()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := db.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead, got:", resp.Nodes[0].State)
	}
}

// Test what happens when a node never starts.
// It should be marked as dead.
func TestNodeHealthInitFail(t *testing.T) {
	ctx := context.Background()
	db, sch := setupSchedTest()

	db.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_INITIALIZING,
	})

	time.Sleep(time.Millisecond * 2)
	sch.CheckNodes()

	resp, err := db.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if resp.Nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead, got:", resp.Nodes[0].State)
	}
}

// Test that a dead node is deleted from the SDB after
// a configurable duration.
func TestNodeHealthDeadTimeout(t *testing.T) {
	ctx := context.Background()
	db, sch := setupSchedTest()

	db.PutNode(ctx, &pbs.Node{
		Id: "test-node",
		// NOTE: the node starts dead
		State: pbs.NodeState_DEAD,
	})

	sch.CheckNodes()
	time.Sleep(time.Millisecond * 2)
	sch.CheckNodes()

	resp, err := db.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Nodes) > 0 {
		t.Error("expected node to be deleted, got", resp.Nodes)
	}
}

func setupSchedTest() (*boltdb.BoltDB, *scheduler.Scheduler) {
	conf := e2e.DefaultConfig()
	conf.Scheduler.NodePingTimeout = time.Millisecond
	conf.Scheduler.NodeInitTimeout = time.Millisecond
	conf.Scheduler.NodeDeadTimeout = time.Millisecond

	back := dummyBackend{}
	db, err := boltdb.NewBoltDB(conf)
	if err != nil {
		panic(err)
	}

	sch := scheduler.NewScheduler(db, back, conf.Scheduler)
	return db, sch
}

type dummyBackend struct{}

func (dummyBackend) GetOffer(t *tes.Task) *scheduler.Offer {
	return nil
}
