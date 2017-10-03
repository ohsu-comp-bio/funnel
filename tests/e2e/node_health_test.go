package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"testing"
	"time"
)

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead
func TestNodeDead(t *testing.T) {
	ctx := context.Background()
	db, sch := setupSchedTest()

	db.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
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

// Test what happens when a node never starts.
// It should be marked as dead.
func TestNodeInitFail(t *testing.T) {
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
func TestNodeDeadTimeout(t *testing.T) {
	ctx := context.Background()
	db, sch := setupSchedTest()

	db.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
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

type dummyBackend struct {
	offerfunc func(*tes.Task) *scheduler.Offer
}

func (d dummyBackend) GetOffer(t *tes.Task) *scheduler.Offer {
	if d.offerfunc != nil {
		return d.offerfunc(t)
	}
	return nil
}

func setupSchedTest() (*boltdb.BoltDB, *scheduler.Scheduler) {
	conf := DefaultConfig()
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
