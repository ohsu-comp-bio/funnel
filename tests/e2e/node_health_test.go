package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"testing"
	"time"
)

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead
func TestNodeDead(t *testing.T) {
	setLogOutput(t)
	conf := nodeTestConfig(DefaultConfig())
	srv := NewFunnel(conf)
	srv.Scheduler = &scheduler.Scheduler{
		DB:      srv.SDB,
		Backend: nil,
		Conf:    conf.Scheduler,
	}

	srv.SDB.PutNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_ALIVE,
	})

	// Some databases need time to sync the PutNode.
	time.Sleep(time.Millisecond * 100)
	// Wait for node to ping timeout.
	time.Sleep(conf.Scheduler.NodePingTimeout)
	// Should mark node as dead.
	srv.Scheduler.CheckNodes()

	nodes := srv.ListNodes()
	if nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead")
	}
}

// Test what happens when a node never starts.
// It should be marked as dead.
func TestNodeInitFail(t *testing.T) {
	setLogOutput(t)
	conf := nodeTestConfig(DefaultConfig())
	srv := NewFunnel(conf)
	srv.Scheduler = &scheduler.Scheduler{
		DB:      srv.SDB,
		Backend: nil,
		Conf:    conf.Scheduler,
	}
	srv.StartServer()

	srv.SDB.PutNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_INITIALIZING,
	})

	time.Sleep(conf.Scheduler.NodeInitTimeout)
	srv.Scheduler.CheckNodes()
	nodes := srv.ListNodes()

	if nodes[0].State != pbs.NodeState_DEAD {
		t.Error("Expected node to be dead")
	}
}

// Test that a dead node is deleted from the SDB after
// a configurable duration.
func TestNodeDeadTimeout(t *testing.T) {
	setLogOutput(t)
	conf := nodeTestConfig(DefaultConfig())
	srv := NewFunnel(conf)
	srv.Scheduler = &scheduler.Scheduler{
		DB:      srv.SDB,
		Backend: nil,
		Conf:    conf.Scheduler,
	}

	srv.SDB.PutNode(context.Background(), &pbs.Node{
		Id:    "test-node",
		State: pbs.NodeState_DEAD,
	})
	srv.Scheduler.CheckNodes()

	time.Sleep(conf.Scheduler.NodeDeadTimeout)
	srv.Scheduler.CheckNodes()
	nodes := srv.ListNodes()

	if len(nodes) > 0 {
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
