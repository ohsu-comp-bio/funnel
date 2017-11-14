package scheduler

import (
	"context"
	workercmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
	"time"
)

// When the node's context is canceled (e.g. because the process
// is being killed) the node should signal the database/server
// that it is gone, and the server will delete the node.
func TestNodeGoneOnCanceledContext(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Backend = "manual"
	conf.Scheduler.NodeInitTimeout = time.Second * 10
	conf.Scheduler.NodePingTimeout = time.Second * 10
	conf.Scheduler.NodeDeadTimeout = time.Second * 10

	srv := e2e.NewFunnel(conf)
	srv.Scheduler = &scheduler.Scheduler{
		DB:      srv.SDB,
		Backend: nil,
		Conf:    conf.Scheduler,
	}
	srv.StartServer()

	srv.Conf.Scheduler.Node.ID = "test-node-gone-on-cancel"
	n, err := scheduler.NewNode(srv.Conf, nil, workercmd.NewDefaultWorker)
	if err != nil {
		t.Fatal("failed to start node")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go n.Run(ctx)

	srv.Scheduler.CheckNodes()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)

	nodes := srv.ListNodes()
	if len(nodes) != 1 {
		t.Fatal("failed to register node", nodes)
	}

	cancel()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)
	srv.Scheduler.CheckNodes()
	nodes = srv.ListNodes()

	if len(nodes) != 0 {
		t.Error("expected node to be deleted")
	}
}

// Run some tasks with the manual backend
func TestManualBackend(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Backend = "manual"
	conf.Scheduler.NodeInitTimeout = time.Second * 10
	conf.Scheduler.NodePingTimeout = time.Second * 10
	conf.Scheduler.NodeDeadTimeout = time.Second * 10

	srv := e2e.NewFunnel(conf)
	srv.Scheduler = &scheduler.Scheduler{
		DB:      srv.SDB,
		Backend: nil,
		Conf:    conf.Scheduler,
	}
	srv.StartServer()

	srv.Conf.Scheduler.Node.ID = "test-node"
	n, err := scheduler.NewNode(srv.Conf, nil, workercmd.NewDefaultWorker)
	if err != nil {
		t.Fatal("failed to start node")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go n.Run(ctx)

	tasks := []string{}
	for i := 1; i <= 10; i++ {
		id := srv.Run(`
      --sh 'echo hello world'
    `)
		tasks = append(tasks, id)
	}

	for _, id := range tasks {
		task := srv.Wait(id)
		if task.State != tes.State_COMPLETE {
			t.Fatal("unexpected task state")
		}

		if task.Logs[0].Logs[0].Stdout != "hello world\n" {
			t.Fatal("Missing stdout")
		}
	}
}
