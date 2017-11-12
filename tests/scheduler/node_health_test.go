package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/tests"
	"testing"
	"time"
)

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead
func TestNodeDead(t *testing.T) {
	ctx := context.Background()
	conf := nodeTestConfig(tests.DefaultConfig())
	srv := tests.NewFunnel(conf)

	_, err := srv.Scheduler.Nodes.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
	})
	if err != nil {
		t.Error(err)
	}

	// Some databases need time to sync the PutNode.
	time.Sleep(time.Millisecond * 100)
	// Wait for node to ping timeout.
	time.Sleep(conf.Scheduler.NodePingTimeout)
	// Should mark node as dead.
	srv.Scheduler.CheckNodes()

	resp, err := srv.Scheduler.Nodes.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	nodes := resp.Nodes

	if len(nodes) < 1 {
		t.Error("expected node was not returned by ListNodes")
	}
	if nodes[0].State != pbs.NodeState_DEAD {
		t.Log("Node:", nodes[0])
		t.Error("Expected node to be dead")
	}
}

// Test what happens when a node never starts.
// It should be marked as dead.
func TestNodeInitFail(t *testing.T) {
	ctx := context.Background()
	conf := nodeTestConfig(tests.DefaultConfig())
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	_, err := srv.Scheduler.Nodes.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_INITIALIZING,
	})
	if err != nil {
		t.Error(err)
	}

	time.Sleep(conf.Scheduler.NodeInitTimeout)
	srv.Scheduler.CheckNodes()

	resp, err := srv.Scheduler.Nodes.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	nodes := resp.Nodes

	if len(nodes) < 1 {
		t.Error("expected node was not returned by ListNodes")
	}
	if nodes[0].State != pbs.NodeState_DEAD {
		t.Log("Node:", nodes[0])
		t.Error("Expected node to be dead")
	}
}

// Test that a dead node is deleted after
// a configurable duration.
func TestNodeDeadTimeout(t *testing.T) {
	ctx := context.Background()
	conf := nodeTestConfig(tests.DefaultConfig())
	srv := tests.NewFunnel(conf)

	_, err := srv.Scheduler.Nodes.PutNode(ctx, &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_DEAD,
	})
	if err != nil {
		t.Error(err)
	}

	srv.Scheduler.CheckNodes()

	time.Sleep(conf.Scheduler.NodeDeadTimeout)
	srv.Scheduler.CheckNodes()

	resp, err := srv.Scheduler.Nodes.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	nodes := resp.Nodes

	if len(nodes) > 0 {
		t.Log("nodes:", nodes)
		t.Error("expected node to be deleted")
	}
}

func nodeTestConfig(conf config.Config) config.Config {
	conf.Backend = "manual"
	conf.Scheduler.NodePingTimeout = time.Millisecond * 300
	conf.Scheduler.NodeInitTimeout = time.Millisecond * 300
	conf.Scheduler.NodeDeadTimeout = time.Millisecond * 300
	return conf
}
