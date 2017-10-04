package node

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/manual"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
	"time"
)

// When the node's context is canceled the node should signal
// the database/server that it is gone, and the server will delete the node.
func TestNodeGoneOnCanceledContext(t *testing.T) {
	conf := e2e.DefaultConfig()
	h := newHarness(conf)
	h.Start()
	defer h.cancel()

	h.sch.CheckNodes()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)

	nodes := h.fun.ListNodes()
	if len(nodes) != 1 {
		t.Fatal("failed to register node", nodes)
	}

	h.cancel()
	time.Sleep(conf.Scheduler.Node.UpdateRate * 2)
	h.sch.CheckNodes()
	nodes = h.fun.ListNodes()

	if len(nodes) != 0 {
		t.Error("expected node to be deleted")
	}
}

// Test that node.Run() exits reasonably soon after canceling the context.
func TestStopNode(t *testing.T) {
	conf := e2e.DefaultConfig()
	h := newHarness(conf)
	h.fun.StartServer()

	// Context used to stop the node.
	ctx, cancel := context.WithCancel(context.Background())

	// Run the node. Add a done channel so we can wait for it to exit.
	done := make(chan struct{})
	go func() {
		h.node.Run(ctx)
		close(done)
	}()

	// Wait for node to initialize.
	time.Sleep(time.Second)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, time.Millisecond*100)
	defer cleanup()

	// Stop the node. Wait for it to exit.
	cancel()
	<-done
}

/*
// Mainly exercising a panic bug caused by an unhandled
// error from client.GetNode().
func TestGetNodeFail(t *testing.T) {
	conf := e2e.DefaultConfig()
  h := newHarness(conf)

	// TODO Set GetNode to return an error

	time.Sleep(time.Second)
  h.cancel()
}
*/

// Test the flow of a node completing a task then timing out.
func TestNodeTimeout(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Scheduler.Node.Timeout = time.Second
	conf.Scheduler.Node.UpdateRate = time.Millisecond * 20

	h := newHarness(conf)
	h.fun.StartServer()

	h.fun.Run(`--sh 'echo hi'`)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, time.Second*2)
	defer cleanup()

	// The node should run, then sit idle until the timeout is reached,
	// at which point Run() exits.
	h.node.Run(context.Background())
	if h.worker.count != 1 {
		t.Error("expected task to be run")
	}
}

// Test that a node does nothing where there are no assigned tasks.
func TestNoTasks(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Scheduler.Node.UpdateRate = time.Millisecond
	h := newHarness(conf)

	h.fun.StartServer()

	// Context used to stop the node.
	ctx, cancel := context.WithCancel(context.Background())

	// Run the node. Add a done channel so we can wait for it to exit.
	done := make(chan struct{})
	go func() {
		h.node.Run(ctx)
		close(done)
	}()

	// Wait for node to initialize.
	time.Sleep(time.Second)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, time.Millisecond*100)
	defer cleanup()

	// Stop the node. Wait for it to exit.
	cancel()
	<-done

	if h.worker.count != 0 {
		t.Fatal("Unexpected worker.Run() call count", h.worker.count)
	}
}

func TestNodeWorkerRun(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Scheduler.Node.UpdateRate = time.Millisecond
	h := newHarness(conf)
	h.Start()

	h.fun.Run(`--sh 'echo hi'`)
	h.fun.Run(`--sh 'echo hi'`)

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, time.Second*10)
	defer cleanup()

	time.Sleep(conf.Scheduler.ScheduleRate + time.Second*3)

	if h.worker.count != 2 {
		t.Fatal("Unexpected worker.Run() call count", h.worker.count)
	}
}

// Counts how many times Run() is called.
type countingWorker struct {
	count int
}

func (c *countingWorker) Run(context.Context, *tes.Task) {
	c.count++
}

type harness struct {
	fun    *e2e.Funnel
	sch    *scheduler.Scheduler
	node   *scheduler.Node
	conf   config.Config
	ctx    context.Context
	worker *countingWorker
	cancel context.CancelFunc
}

func (h *harness) Start() {
	h.fun.StartServer()
	go h.node.Run(h.ctx)
}

func newHarness(conf config.Config) *harness {
	h := harness{}
	h.ctx, h.cancel = context.WithCancel(context.Background())

	bak, err := manual.NewBackend(conf)
	if err != nil {
		panic(err)
	}

	h.fun = e2e.NewFunnel(conf)
	h.sch = scheduler.NewScheduler(h.fun.SDB, bak, conf.Scheduler)
	h.fun.Scheduler = h.sch

	h.fun.DB.WithComputeBackend(scheduler.NewComputeBackend(h.fun.SDB))

	h.worker = &countingWorker{}
	n, err := scheduler.NewNode(conf, h.worker)
	h.node = n
	if err != nil {
		panic(err)
	}
	return &h
}

func timeLimit(t *testing.T, d time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		select {
		case <-time.NewTimer(d).C:
			t.Fatal("time limit expired")
		case <-stop:
		}
	}()
	return func() {
		close(stop)
	}
}
