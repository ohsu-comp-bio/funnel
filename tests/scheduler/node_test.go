package scheduler

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	workercmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
)

// When the node's context is canceled (e.g. because the process
// is being killed) the node should signal the database/server
// that it is gone, and the server will delete the node.
func TestNodeGoneOnCanceledContext(t *testing.T) {
	conf := tests.DefaultConfig()
	conf.Compute = "manual"
	conf.Scheduler.NodeInitTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodePingTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodeDeadTimeout = config.Duration(time.Second * 10)

	bg := context.Background()
	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	srv.Conf.Node.ID = "test-node-gone-on-cancel"
	n, err := scheduler.NewNodeProcess(bg, srv.Conf, scheduler.NoopWorker, log)
	if err != nil {
		t.Fatal("failed to start node", err)
	}
	ctx, cancel := context.WithCancel(bg)
	defer cancel()
	go n.Run(ctx)

	srv.Scheduler.CheckNodes()
	time.Sleep(time.Duration(conf.Node.UpdateRate * 2))

	resp, err := srv.Scheduler.Nodes.ListNodes(bg, &scheduler.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	nodes := resp.Nodes

	if len(nodes) != 1 {
		t.Fatal("failed to register node", nodes)
	}

	cancel()
	time.Sleep(time.Duration(conf.Node.UpdateRate * 2))
	srv.Scheduler.CheckNodes()

	resp, err = srv.Scheduler.Nodes.ListNodes(bg, &scheduler.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	nodes = resp.Nodes

	if len(nodes) != 0 {
		t.Error("expected node to be deleted")
	}
}

// Run some tasks with the manual backend
func TestManualBackend(t *testing.T) {
	conf := tests.DefaultConfig()
	conf.Compute = "manual"
	conf.Scheduler.NodeInitTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodePingTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodeDeadTimeout = config.Duration(time.Second * 10)

	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv.Conf.Node.ID = "test-node-manual"
	// create a node
	srv.Conf.Node.ID = "test-node-manual"
	w, err := workercmd.NewWorker(ctx, conf, log)
	if err != nil {
		t.Fatal("failed to create worker factory", err)
	}
	n, err := scheduler.NewNodeProcess(ctx, srv.Conf, w.Run, log)
	if err != nil {
		t.Fatal("failed to create node", err)
	}

	go n.Run(ctx)

	// run tasks and check that they all complete
	tasks := []string{}
	for i := 0; i < 10; i++ {
		id := srv.Run(`
      --sh 'echo hello world'
    `)
		tasks = append(tasks, id)
	}

	for _, id := range tasks {
		task := srv.Wait(id)
		time.Sleep(time.Millisecond * 100)
		if task.State != tes.State_COMPLETE {
			t.Fatal("unexpected task state")
		}

		if task.Logs[0].Logs[0].Stdout != "hello world\n" {
			t.Fatalf("Missing stdout for task %s", id)
		}
	}
}

func TestDeadNodeTaskCleanup(t *testing.T) {
	conf := tests.DefaultConfig()
	conf.Compute = "manual"
	conf.Scheduler.NodeInitTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodePingTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodeDeadTimeout = config.Duration(time.Second * 10)

	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	srv.Conf.Node.ID = "test-node-manual"
	blockingNoopWorker := func(ctx context.Context, taskID string) error {
		time.Sleep(time.Minute * 10)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	n, err := scheduler.NewNodeProcess(ctx, srv.Conf, blockingNoopWorker, log)
	if err != nil {
		t.Fatal("failed to create node")
	}
	go n.Run(ctx)

	id := srv.Run(`
      --sh 'echo hello world'
  `)

	// wait for the task to be assigned to a node
	srv.WaitForInitializing(id)

	// cancel context of Node
	cancel()

	// scheduler should discover node is gone and update the task state accordingly
	task := srv.Wait(id)
	if task.State != tes.State_SYSTEM_ERROR {
		t.Fatal("unexpected task state")
	}
}

// Tests a bug where tasks and nodes were not being correctly cleaned up
// when the node crashed and was restarted.
func TestNodeCleanup(t *testing.T) {
	log := logger.NewLogger("node", tests.LogConfig())
	ctx := context.Background()

	conf := tests.DefaultConfig()
	conf.Compute = "manual"
	srv := tests.NewFunnel(conf)

	e := srv.Server.Events

	t1 := tests.HelloWorld()
	srv.Server.Tasks.CreateTask(ctx, t1)
	e.WriteEvent(ctx, events.NewState(t1.Id, tes.Complete))

	t2 := tests.HelloWorld()
	srv.Server.Tasks.CreateTask(ctx, t2)
	e.WriteEvent(ctx, events.NewState(t2.Id, tes.Running))

	t3 := tests.HelloWorld()
	srv.Server.Tasks.CreateTask(ctx, t3)
	e.WriteEvent(ctx, events.NewState(t3.Id, tes.SystemError))

	t4 := tests.HelloWorld()
	srv.Server.Tasks.CreateTask(ctx, t4)
	e.WriteEvent(ctx, events.NewState(t4.Id, tes.Running))

	t5 := tests.HelloWorld()
	srv.Server.Tasks.CreateTask(ctx, t5)
	e.WriteEvent(ctx, events.NewState(t5.Id, tes.Running))

	srv.Scheduler.Nodes.PutNode(ctx, &scheduler.Node{
		Id:      "test-gone-node-cleanup-restart-1",
		State:   scheduler.NodeState_GONE,
		TaskIds: []string{t1.Id, t2.Id, t3.Id},
	})

	srv.Scheduler.Nodes.PutNode(ctx, &scheduler.Node{
		Id:      "test-gone-node-cleanup-restart-2",
		State:   scheduler.NodeState_GONE,
		TaskIds: []string{t4.Id},
	})

	srv.Scheduler.Nodes.PutNode(ctx, &scheduler.Node{
		Id:      "test-gone-node-cleanup-restart-3",
		State:   scheduler.NodeState_ALIVE,
		TaskIds: []string{t5.Id},
	})

	ns, _ := srv.Scheduler.Nodes.ListNodes(ctx, &scheduler.ListNodesRequest{})
	log.Info("nodes before", ns)

	err := srv.Scheduler.CheckNodes()
	if err != nil {
		t.Error(err)
	}

	ns, _ = srv.Scheduler.Nodes.ListNodes(ctx, &scheduler.ListNodesRequest{})
	if len(ns.Nodes) != 1 {
		t.Error("expected 1 node")
	}

	if ns.Nodes[0].Id != "test-gone-node-cleanup-restart-3" {
		t.Error("unexpected node")
	}

	ts, _ := srv.Server.Tasks.ListTasks(ctx, &tes.ListTasksRequest{})
	if len(ts.Tasks) != 5 {
		log.Info("tasks", ts)
		t.Error("expected 5 tasks")
	}

	expected := []tes.State{
		tes.Running,
		tes.SystemError,
		tes.SystemError,
		tes.SystemError,
		tes.Complete,
	}

	for i, task := range ts.Tasks {
		e := expected[i]
		if task.State != e {
			t.Error("expected state for task", i, task.State, e)
		}
	}
}

// Test that nodes may have a configured zone, and that tasks may request a zone,
// and that tasks are scheduled properly based on their requested zone.
func TestNodeZoneScheduling(t *testing.T) {

	//////////////////////// Setup

	conf := tests.DefaultConfig()
	conf.Compute = "manual"
	conf.Scheduler.ScheduleChunk = 50
	conf.Scheduler.NodeInitTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodePingTimeout = config.Duration(time.Second * 10)
	conf.Scheduler.NodeDeadTimeout = config.Duration(time.Second * 10)

	bg := context.Background()
	ctx, cancel := context.WithCancel(bg)
	defer cancel()

	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)

	srv := tests.NewFunnel(conf)
	srv.StartServer()

	// Set up a node in "zone-1"

	n1conf := tests.DefaultConfig()
	n1conf.Server = srv.Conf.Server
	n1conf.Node.ID = "test-node-1"
	n1conf.Node.Resources.Zone = "zone-1"
	n1conf.Node.Resources.Cpus = 20

	n1, err := scheduler.NewNodeProcess(ctx, n1conf, scheduler.NoopWorker, log)
	if err != nil {
		t.Fatal("failed to start node", err)
	}
	go n1.Run(ctx)

	// Set up a node in "zone-2"

	n2conf := tests.DefaultConfig()
	n2conf.Server = srv.Conf.Server
	n2conf.Node.ID = "test-node-2"
	n2conf.Node.Resources.Zone = "zone-2"
	n2conf.Node.Resources.Cpus = 20

	n2, err := scheduler.NewNodeProcess(ctx, n2conf, scheduler.NoopWorker, log)
	if err != nil {
		t.Fatal("failed to start node", err)
	}
	go n2.Run(ctx)

	///////////////////// Tests

	// Create 10 tasks for "zone-1" and 10 tasks for "zone-2"
	// Create 10 tasks for "zone-3" which doesn't exist
	var zone1ids, zone2ids []string
	for i := 0; i < 10; i++ {
		id1 := srv.Run(`--sh "echo" --cpu 1 --zone zone-1`)
		zone1ids = append(zone1ids, id1)
		id2 := srv.Run(`--sh "echo" --cpu 1 --zone zone-2`)
		zone2ids = append(zone2ids, id2)
		srv.Run(`--sh "echo" --cpu 1 --zone zone-3`)
	}

	// The scheduler databases sometimes fail on version conflicts.
	// In normal operation, the tasks scheduling is retried on the next iteration.
	// Fake that here.
	for i := 0; i < 5; i++ {
		srv.Scheduler.Schedule(ctx)
	}

	resp, err := srv.Scheduler.Nodes.ListNodes(bg, &scheduler.ListNodesRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Nodes) != 2 {
		t.Fatal("expected two nodes")
	}

	for _, node := range resp.Nodes {
		var expectedIDs []string
		switch node.Id {
		case "test-node-1":
			expectedIDs = zone1ids
		case "test-node-2":
			expectedIDs = zone2ids
		default:
			t.Fatal("unexpected node ID", node.Id)
		}

		sort.Strings(node.TaskIds)
		if !reflect.DeepEqual(node.TaskIds, expectedIDs) {
			t.Error("unexpected node task IDs", node.TaskIds, expectedIDs)
		}
	}
}
