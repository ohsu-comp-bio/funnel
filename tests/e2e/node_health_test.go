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

	srv.DB.UpdateNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
	})

	time.Sleep(conf.Scheduler.NodePingTimeout * 2)
	srv.DB.CheckNodes()

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

	srv.DB.UpdateNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_INITIALIZING,
	})

	time.Sleep(conf.Scheduler.NodeInitTimeout * 2)
	srv.DB.CheckNodes()
	nodes := srv.ListNodes()

	if nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead")
	}
}
