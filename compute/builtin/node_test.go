package builtin

import (
  "context"
	"testing"
	"time"
  "sync"

	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Test calling stopping a node by canceling its context
func TestStopNode(t *testing.T) {
	conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, 200*time.Millisecond)
	defer cleanup()

	cancel := n.Start()
	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

	// Cancel the node's context.
	cancel()

	// Wait for node to finish.
	n.Wait()

	// Give the scheduler time to disconnect.
	time.Sleep(30 * time.Millisecond)

	h, ok := s.handles[n.detail.Id]
	if !ok {
		t.Fatal("didn't find node record")
	}
	if h.node.State != NodeState_GONE {
		t.Errorf("expected state to be GONE, but got %s", h.node.State)
	}
}

// Test that a disconnected (connection dropped, not node stopped)
// node is marked as dead.
func TestNodeDisconnected(t *testing.T) {
	conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, 200*time.Millisecond)
	defer cleanup()

	n.Start()
	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

	// Force close the grpc connection without properly shutting down the node.
	n.conn.Close()

	// Give the scheduler time to disconnect.
	time.Sleep(30 * time.Millisecond)

	h, ok := s.handles[n.detail.Id]
	if !ok {
		t.Fatal("didn't find node record")
	}
	if h.node.State != NodeState_DEAD {
		t.Errorf("expected state to be DEAD, but got %s", h.node.State)
	}
}

// Test the simple case of a node that is alive,
// then doesn't ping in time, and it marked dead.
//
// This is a legacy test and makes less sense with streaming code;
// since the scheduler should detect disconnects immediately.
// But, it's easy to test and maybe could happen.
func TestNodeDead(t *testing.T) {
	ctx := context.Background()
	conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

	time.Sleep(20 * time.Millisecond)

  // Connect and send a single ping to register the node.
  n.connect(ctx)
  err := n.ping()
  if err != nil {
    t.Fatal(err)
  }

	time.Sleep(20 * time.Millisecond)

	h, ok := s.handles[n.detail.Id]
	if !ok {
		t.Fatal("didn't find node record")
	}
	if h.node.State != NodeState_ALIVE {
		t.Errorf("expected state to be ALIVE, but got %s", h.node.State)
	}

	// Wait for node to ping timeout.
	time.Sleep(time.Duration(conf.Scheduler.NodePingTimeout))

  // Check the nodes.
  s.checkNodes(ctx)

	h, ok = s.handles[n.detail.Id]
	if !ok {
		t.Fatal("didn't find node record")
	}
	if h.node.State != NodeState_DEAD {
		t.Errorf("expected state to be DEAD, but got %s", h.node.State)
	}
}

// Test that a node that doesn't ping is marked dead after some time limit.
func TestNodePingTimeout(t *testing.T) {
	conf := testConfig()
	s := newTestSched(conf)

  // This node should be dead, its LastPing is greater than the timeout.
  t1 := time.Duration(conf.Scheduler.NodePingTimeout)
  h1 := &nodeHandle{
    node: &Node{
      State: NodeState_ALIVE,
      LastPing: time.Now().Add(-t1).UnixNano(),
    },
  }

  // This node should be alive, its LastPing timeout is recent enough.
  t2 := time.Duration(conf.Scheduler.NodePingTimeout) - (20 * time.Millisecond)
  h2 := &nodeHandle{
    node: &Node{
      State: NodeState_ALIVE,
      LastPing: time.Now().Add(-t2).UnixNano(),
    },
  }

  s.handles["node-1"] = h1
  s.handles["node-2"] = h2
	ctx := context.Background()
  s.checkNodes(ctx)

  if h1.node.State != NodeState_DEAD {
    t.Errorf("expected node-1 to be DEAD, but got %s", h1.node.State)
  }

  if h2.node.State != NodeState_ALIVE {
    t.Errorf("expected node-2 to be ALIVE, but got %s", h2.node.State)
  }
}

// Test that dead nodes are cleanup up after some time.
func TestDeadNodeCleanedUp(t *testing.T) {
	conf := testConfig()
	s := newTestSched(conf)

  // This node should be removed.
  t1 := time.Duration(conf.Scheduler.NodeDeadTimeout)
  h1 := &nodeHandle{
    node: &Node{
      State: NodeState_DEAD,
      LastPing: time.Now().Add(-t1).UnixNano(),
    },
  }

  // This node should remain.
  t2 := time.Duration(conf.Scheduler.NodeDeadTimeout) - (20 * time.Millisecond)
  h2 := &nodeHandle{
    node: &Node{
      State: NodeState_DEAD,
      LastPing: time.Now().Add(-t2).UnixNano(),
    },
  }

  s.handles["node-1"] = h1
  s.handles["node-2"] = h2
	ctx := context.Background()
  s.checkNodes(ctx)

  _, ok := s.handles["node-1"]
  if ok {
    t.Error("expected node-1 to be removed")
  }

  _, ok = s.handles["node-2"]
  if !ok {
    t.Error("node-2 was removed unexpectedly")
  }
}

// Test that a worker gets created for each task.
func TestNodeWorkerCreated(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

  n.Start()

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

	// Count the number of times the worker factory was called
	var count int
	n.workerRun = func(context.Context, string) error {
		count++
		return nil
	}

  s.scheduleOne(&tes.Task{Id: "task-1"})
  s.scheduleOne(&tes.Task{Id: "task-2"})

	time.Sleep(20 * time.Millisecond)

	if count != 2 {
		t.Fatalf("Unexpected node worker count: %d", count)
	}
}

func TestGetNode(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

  n.Start()

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

  ctx := context.Background()
  _, err := s.GetNode(ctx, &GetNodeRequest{
    Id: n.detail.Id,
  })
  if err != nil {
    t.Fatal(err)
  }
}

func TestGetNodeMissing(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

  ctx := context.Background()
  _, err := s.GetNode(ctx, &GetNodeRequest{
    Id: "unknown",
  })
	if status.Code(err) != codes.NotFound {
    t.Errorf("expected not found error, but got %s", err)
  }
}

func TestListNodes(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)
	n2 := newTestNode(conf)

  n.Start()
  n2.Start()

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

  ctx := context.Background()
  resp, err := s.ListNodes(ctx, &ListNodesRequest{})
  if err != nil {
    t.Fatal(err)
  }

  if len(resp.Nodes) != 2 {
    t.Error("expected 2 nodes")
  }
}

func TestListNodesEmpty(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)

  ctx := context.Background()
  resp, err := s.ListNodes(ctx, &ListNodesRequest{})
  if err != nil {
    t.Fatal(err)
  }

  if len(resp.Nodes) != 0 {
    t.Error("expected 0 nodes")
  }
}

func TestDrainNode(t *testing.T) {
  ctx := context.Background()
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

  n.Start()

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

  _, err := s.DrainNode(ctx, &DrainNodeRequest{
    Id: n.detail.Id,
  })
  if err != nil {
    t.Fatal(err)
  }

	time.Sleep(50 * time.Millisecond)

  resp, err := s.GetNode(ctx, &GetNodeRequest{
    Id: n.detail.Id,
  })
  if err != nil {
    t.Fatal(err)
  }

  if resp.State != NodeState_DRAIN {
		t.Errorf("expected state to be DRAIN, but got %s", resp.State)
  }
}


// Test that a panic from the worker doesn't crash the node.
func TestNodeWorkerPanic(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)

  n.Start()

	// Give the scheduler server time to start.
	time.Sleep(20 * time.Millisecond)

	// Count the number of times the worker factory was called
	n.workerRun = func(context.Context, string) error {
    panic("test panic")
	}

  s.scheduleOne(&tes.Task{Id: "task-1"})

	time.Sleep(20 * time.Millisecond)

	h, ok := s.handles[n.detail.Id]
	if !ok {
		t.Fatal("didn't find node record")
	}
	if h.node.State != NodeState_ALIVE {
		t.Errorf("expected state to be ALIVE, but got %s", h.node.State)
	}
}

func TestNodeDetectResources(t *testing.T) {
  conf := testConfig()
  conf.Node.Resources.Cpus = 1234
  conf.Node.Resources.RamGb = 1235.0
  conf.Node.Resources.DiskGb = 1236.0
	s := newTestSched(conf)
	n := newTestNode(conf)
  n.Start()

	n.workerRun = func(context.Context, string) error {
    time.Sleep(10 *time.Second)
		return nil
	}

	// Give the scheduler server time to start.
	time.Sleep(50 * time.Millisecond)

  d := n.detail.Resources
  if d.Cpus != 1234 {
    t.Errorf("expected 1234 Cpus, got %d", d.Cpus)
  }
  if d.RamGb != 1235.0 {
    t.Errorf("expected 1235 RamGb, got %f", d.RamGb)
  }
  if d.DiskGb != 1236.0 {
    t.Errorf("expected 1236 Cpus, got %f", d.DiskGb)
  }

  // Check that the scheduler has the same resources
  d = s.handles[n.detail.Id].node.Resources
  if d.Cpus != 1234 {
    t.Errorf("expected 1234 Cpus, got %d", d.Cpus)
  }
  if d.RamGb != 1235.0 {
    t.Errorf("expected 1235 RamGb, got %f", d.RamGb)
  }
  if d.DiskGb != 1236.0 {
    t.Errorf("expected 1236 Cpus, got %f", d.DiskGb)
  }

  // !!!!!!!!!!!!
  // Note that tasks have a minimum cpu resource request of 1 CPU.

  s.assignTask(&tes.Task{Id: "task-1"}, n.detail.Id)
	time.Sleep(50 * time.Millisecond)

  d = n.detail.Available
  if d.Cpus != 1233 {
    t.Errorf("expected 1234 Cpus, got %d", d.Cpus)
  }
  if d.RamGb != 1235.0 {
    t.Errorf("expected 1235 RamGb, got %f", d.RamGb)
  }
  if d.DiskGb != 1236.0 {
    t.Errorf("expected 1236 Cpus, got %f", d.DiskGb)
  }

  // Check that the scheduler has the same resources
  d = s.handles[n.detail.Id].node.Available
  if d.Cpus != 1233 {
    t.Errorf("expected 1234 Cpus, got %d", d.Cpus)
  }
  if d.RamGb != 1235.0 {
    t.Errorf("expected 1235 RamGb, got %f", d.RamGb)
  }
  if d.DiskGb != 1236.0 {
    t.Errorf("expected 1236 Cpus, got %f", d.DiskGb)
  }
}

// Test that a node does nothing where there are no assigned tasks.
func TestNodeWorkerNoTasks(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)
	n2 := newTestNode(conf)

  n.Start()
  n2.Start()

	// Give the scheduler server time to start.
	time.Sleep(50 * time.Millisecond)

	// Count the number of times the worker factory was called
  wg := sync.WaitGroup{}
	var count, count2 int
	n.workerRun = func(context.Context, string) error {
		count++
    wg.Done()
		return nil
	}
	n2.workerRun = func(context.Context, string) error {
		count2++
    wg.Done()
		return nil
	}

  wg.Add(2)
  s.assignTask(&tes.Task{Id: "task-1"}, n.detail.Id)
  s.assignTask(&tes.Task{Id: "task-2"}, n.detail.Id)

  wg.Wait()

	if count != 2 {
		t.Fatalf("Expected worker count 2, got %d", count)
	}
	if count2 != 0 {
		t.Fatalf("Expected worker count 0, got %d", count2)
	}
}

// TODO test:
// - that detect resources config overrides autodetection properly
// - that available resources are calculated correctly
// - that resources have a minimum request, e.g. cpu

/*

func TestNoTasks(t *testing.T) {
  conf := testConfig()
	n := newTestNode(conf, t)

	// Tell the scheduler mock to return nothing
	n.Client.On("GetNode", mock.Anything, mock.Anything, mock.Anything).
		Return(&Node{}, nil)

	// Count the number of times the worker factory was called
	var count int
	n.workerRun = func(context.Context, string) error {
		count++
		return nil
	}

	n.sync(context.Background())
	n.sync(context.Background())
	n.sync(context.Background())
	time.Sleep(time.Second)

	if count != 0 {
		t.Fatal("Unexpected worker factory call count")
	}
	if n.workers.Count() != 0 {
		t.Fatal("Unexpected node worker count")
	}
}

// Test that a finished task is not immediately re-run.
// Tests a bugfix.
func TestFinishedTaskNotRerun(t *testing.T) {
  conf := testConfig()
	n := newTestNode(conf, t)

	// Set up a test worker which this code can easily control.
	//w := testWorker{}
	// Hook the test worker up to the node's worker factory.
	//n.newWorker = Worker(w.Factory)

	n.AddTasks("task-1")

	// manually sync the node to avoid timing issues.
	n.sync(context.Background())
	time.Sleep(time.Second)

	if n.workers.Count() != 0 {
		t.Fatalf("Unexpected worker count: %d", n.workers.Count())
	}

	// There was a bug where later syncs would end up re-running the task.
	// Do a few syncs to make sure.
	n.sync(context.Background())
	n.sync(context.Background())
	time.Sleep(time.Second)

	if n.workers.Count() != 0 {
		t.Fatalf("Unexpected worker count: %d", n.workers.Count())
	}
}

// Test that tasks are removed from the node's runset when they finish.
func TestFinishedTaskRunsetCount(t *testing.T) {
  conf := testConfig()
	n := newTestNode(conf, t)

	// Set up a test worker which this code can easily control.
	//w := testWorker{}
	// Hook the test worker up to the node's worker factory.
	//n.newWorker = Worker(w.Factory)

	n.AddTasks("task-1")

	// manually sync the node to avoid timing issues.
	n.sync(context.Background())
	time.Sleep(time.Second)

	if n.workers.Count() != 0 {
		t.Fatalf("Unexpected worker count: %d", n.workers.Count())
	}
}

// Run some tasks with the builtin backend
func TestBuiltinBackend(t *testing.T) {
  conf := testConfig()

	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv.Conf.Node.ID = "test-node-builtin"
	// create a node
	srv.Conf.Node.ID = "test-node-builtin"
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
  conf := testConfig()

	log := logger.NewLogger("node", tests.LogConfig())
	tests.SetLogOutput(log, t)
	srv := tests.NewFunnel(conf)
	srv.StartServer()

	srv.Conf.Node.ID = "test-node-builtin"
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
  conf := testConfig()

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
*/
