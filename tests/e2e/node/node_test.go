package node

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
	"time"
)

// When the node's context is canceled (e.g. because the process
// is being killed) the node should signal the database/server
// that it is gone, and the server will delete the node.
func TestNodeGoneOnCanceledContext(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Scheduler.NodeInitTimeout = time.Second * 10
	conf.Scheduler.NodePingTimeout = time.Second * 10
	conf.Scheduler.NodeDeadTimeout = time.Second * 10

	srv := e2e.NewFunnel(conf)
	srv.StartServer()

	srv.Conf.Scheduler.Node.ID = "test-node-gone-on-cancel"
	n, err := scheduler.NewNode(srv.Conf)
	if err != nil {
		t.Fatal("failed to start node")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go n.Run(ctx)

	srv.DB.CheckNodes()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)

	nodes := srv.ListNodes()
	if len(nodes) != 1 {
		t.Fatal("failed to register node", nodes)
	}

	cancel()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)
	srv.DB.CheckNodes()
	nodes = srv.ListNodes()

	if len(nodes) != 0 {
		t.Error("expected node to be deleted")
	}
}
