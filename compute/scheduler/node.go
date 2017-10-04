package scheduler

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/rpc"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"golang.org/x/sync/syncmap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"sync"
	"time"
)

// NewNode returns a new Node instance
func NewNode(conf config.Config, w worker.Worker) (*Node, error) {
	log := logger.Sub("node", "nodeID", conf.Scheduler.Node.ID)

	// Detect available resources at startup
	res := detectResources(conf.Scheduler.Node)
	timeout := util.NewIdleTimeout(conf.Scheduler.Node.Timeout)
	state := pbs.NodeState_UNINITIALIZED
	hostname, _ := os.Hostname()

	log.Debug("config", "node", conf.Scheduler.Node)

	return &Node{
		conf:      conf.Scheduler.Node,
		log:       log,
		resources: res,
		worker:    w,
		timeout:   timeout,
		state:     state,
		hostname:  hostname,
	}, nil
}

// Node is a structure used for tracking available resources on a compute resource.
type Node struct {
	conf      config.Node
	client    Client
	log       logger.Logger
	resources pbs.Resources
	tesc      *rpc.TESClient
	worker    worker.Worker
	tasks     syncmap.Map
	timeout   util.IdleTimeout
	state     pbs.NodeState
	busy      sync.WaitGroup
	taskCount int
	hostname  string
}

// Run runs a node with the given config. This is responsible for communication
// with the server and starting task workers
func (n *Node) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cli, err := NewClient(n.conf.RPC)
	if err != nil {
		return fmt.Errorf("node can't connect sched client %s", err)
	}
	n.client = cli

	tesc, err := rpc.NewTESClient(n.conf.RPC)
	if err != nil {
		return fmt.Errorf("node can't connect TES client %s", err)
	}
	n.tesc = tesc

	n.log.Info("Starting node")
	n.state = pbs.NodeState_ALIVE
	n.checkConnection(ctx)
	n.sync(ctx)

	ticker := time.NewTicker(n.conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		case <-n.timeout.Done():
			cancel()
		case <-ctx.Done():
			n.timeout.Stop()

			// The node gets 10 seconds to do a final sync with the scheduler.
			stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			n.state = pbs.NodeState_GONE
			n.sync(stopCtx)
			n.client.Close()

			// The tasks get 10 seconds to finish up.
			limitedWait(time.Second*10, &n.busy)
			return ctx.Err()
		case <-ticker.C:
			n.sync(ctx)
			n.checkIdleTimer()
		}
	}
}

func limitedWait(dur time.Duration, wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(dur):
		return fmt.Errorf("node stop timed out")
	}
}

func (n *Node) checkConnection(ctx context.Context) {
	_, err := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})

	// If its a 404 error create a new node
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		log.Error("Couldn't contact server.", err)
	} else {
		log.Info("Successfully connected to server.")
	}
}

func (n *Node) runTask(ctx context.Context, task *tes.Task) {
	n.log.Debug("Starting worker for task", task.Id)

	tctx := pollForCancel(ctx, task.Id, n.tesc)
	n.worker.Run(tctx, task)
	// Make sure the task is cleaned up from the node's task map.
	defer n.tasks.Delete(task.Id)

	log.Debug("AFTER")
	// task cannot fully complete until it has successfully
	// removed the assigned task ID from the database.
	ticker := time.NewTicker(n.conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug("UPDATE TICK")
			r, err := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})
			if err != nil {
				log.Error("Couldn't get node state during task sync.", err)
				// Break out of "select" statement, but not "for" loop.
				break
			}

			// Find the finished task ID and remove it from the node record.
			for i, id := range r.TaskIds {
				if id == task.Id {
					r.TaskIds = append(r.TaskIds[:i], r.TaskIds[i+1:]...)
					break
				}
			}
			log.Debug("UPDATE", r)

			_, err = n.client.PutNode(context.Background(), r)
			if err != nil {
				log.Error("Couldn't save node state during task sync.", err)
				// Break out of "select" statement, but not "for" loop.
				break
			}
			// Update was a success, so return.
			return
		}
	}
}

// sync syncs the node's state with the server. It reports task state changes,
// handles signals from the server (new task, cancel task, etc), reports resources, etc.
//
// TODO Sync should probably use a channel to sync data access.
//      Probably only a problem for test code, where Sync is called directly.
func (n *Node) sync(ctx context.Context) {

	r, err := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})
	if err != nil {
		// If its a 404 error create a new node
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			log.Error("Couldn't get node state during sync.", err)
			return
		}
		log.Info("Starting initial node sync", "nodeID", n.conf.ID)
	}

	// Start task workers. "n.tasks" will track task IDs
	// to ensure there's only one worker per ID.
	for _, id := range r.GetTaskIds() {
		// Check if the task is already running. If not, start a worker.
		if _, ok := n.tasks.Load(id); !ok {
			// Get the full task and store it
			task, err := n.tesc.FullTask(id)
			if err != nil {
				log.Error("error retrieving task. skipping", id)
				continue
			}
			n.tasks.Store(id, task)
			go n.runTask(ctx, task)
		}
	}

	// Collect all the running tasks in order to update available resources,
	// update task count, and sync active task IDs with the database.
	var tasks []*tes.Task
	var ids []string

	n.tasks.Range(func(key, val interface{}) bool {
		id := key.(string)
		task := val.(*tes.Task)
		tasks = append(tasks, task)
		ids = append(ids, id)
		// continue iteration
		return true
	})
	n.taskCount = len(tasks)

	// Updatind disk space is tricky because it's constantly changing.
	// The simple approach taken here is to only change it when the node
	// is idle. Probably want something smarter at some point.
	if n.taskCount == 0 {
		n.resources = detectResources(n.conf)
	}
	a := AvailableResources(tasks, &n.resources)

	// Merge metadata
	m := map[string]string{}
	for k, v := range n.conf.Metadata {
		m[k] = v
	}
	for k, v := range r.GetMetadata() {
		m[k] = v
	}

	_, err = n.client.PutNode(context.Background(), &pbs.Node{
		Id:        n.conf.ID,
		Resources: &n.resources,
		Available: a,
		State:     n.state,
		Hostname:  n.hostname,
		Version:   r.GetVersion(),
		Metadata:  n.conf.Metadata,
		TaskIds:   ids,
	})
	if err != nil {
		log.Error("couldn't save node update", err)
		return
	}
}

// Check if the node is idle. If so, start the timeout timer.
func (n *Node) checkIdleTimer() {
	// The pool is idle if there are no task workers.
	// The pool should not time out if it's not alive (e.g. if it's initializing)
	idle := n.taskCount == 0 && n.state == pbs.NodeState_ALIVE
	if idle {
		n.timeout.Start()
	} else {
		n.timeout.Stop()
	}
}

func pollForCancel(ctx context.Context, id string, c *rpc.TESClient) context.Context {
	taskctx, cancel := context.WithCancel(ctx)

	// Start a goroutine that polls the server to watch for a canceled state.
	// If a cancel state is found, "taskctx" is canceled.
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-taskctx.Done():
				return
			case <-ticker.C:
				state, _ := c.State(id)
				if tes.TerminalState(state) {
					cancel()
				}
			}
		}
	}()
	return taskctx
}
