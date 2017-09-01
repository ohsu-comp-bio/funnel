package e2e

import (
	"context"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"testing"
	"time"
)

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead
func TestNodeDead(t *testing.T) {
	conf := DefaultConfig()
	conf.Scheduler.NodePingTimeout = time.Millisecond
	srv := NewFunnel(conf)

	srv.SDB.UpdateNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
	})

	time.Sleep(conf.Scheduler.NodePingTimeout * 2)
	srv.SDB.CheckNodes()

	nodes := srv.ListNodes()
	if nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead")
	}
}

// Test what happens when a node never starts.
// It should be marked as dead.
func TestNodeInitFail(t *testing.T) {
	conf := DefaultConfig()
	conf.Scheduler.NodeInitTimeout = time.Millisecond
	srv := NewFunnel(conf)
	srv.StartServer()

	srv.SDB.UpdateNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_INITIALIZING,
	})

	time.Sleep(conf.Scheduler.NodeInitTimeout * 2)
	srv.SDB.CheckNodes()
	nodes := srv.ListNodes()

	if nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead")
	}
}

// Test that a dead node is deleted from the SDB after
// a configurable duration.
func TestNodeDeadTimeout(t *testing.T) {
	conf := DefaultConfig()
	conf.Scheduler.NodeInitTimeout = time.Millisecond
	conf.Scheduler.NodeDeadTimeout = time.Millisecond
	srv := NewFunnel(conf)

	srv.SDB.UpdateNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_DEAD,
	})
	srv.SDB.CheckNodes()

	time.Sleep(conf.Scheduler.NodeDeadTimeout * 2)
	srv.SDB.CheckNodes()
	nodes := srv.ListNodes()

	if len(nodes) > 0 {
		t.Error("expected node to be deleted")
	}
}
