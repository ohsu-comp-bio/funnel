package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"time"
)

// NewNode returns a new Node instance
func NewNode(conf config.Config) (*Node, error) {
	log := logger.Sub("node", "nodeID", conf.Scheduler.Node.ID)

	cli, err := NewClient(conf.Scheduler)
	if err != nil {
		return nil, err
	}

	err = util.EnsureDir(conf.Scheduler.Node.WorkDir)
	if err != nil {
		return nil, err
	}

	// Detect available resources at startup
	res := detectResources(conf.Scheduler.Node)
	timeout := util.NewIdleTimeout(conf.Scheduler.Node.Timeout)
	state := pbs.NodeState_UNINITIALIZED

	workerConf := conf.Worker
	workerConf.WorkDir = conf.Scheduler.Node.WorkDir

	log.Debug("config", "node", conf.Scheduler.Node, "worker", workerConf)

	return &Node{
		conf:       conf.Scheduler.Node,
		workerConf: workerConf,
		client:     cli,
		log:        log,
		resources:  res,
		newWorker:  worker.NewDefaultWorker,
		workers:    newRunSet(),
		timeout:    timeout,
		state:      state,
	}, nil
}

// NewNoopNode returns a new node that doesn't have any side effects
// (e.g. storage access, docker calls, etc.) which is useful for testing.
func NewNoopNode(conf config.Config) (*Node, error) {
	n, err := NewNode(conf)
	n.newWorker = NoopWorkerFactory
	return n, err
}

// Node is a structure used for tracking available resources on a compute resource.
type Node struct {
	conf       config.Node
	workerConf config.Worker
	client     Client
	log        logger.Logger
	resources  pbs.Resources
	newWorker  WorkerFactory
	workers    *runSet
	timeout    util.IdleTimeout
	state      pbs.NodeState
}

// Run runs a node with the given config. This is responsible for communication
// with the server and starting task workers
func (n *Node) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	n.log.Info("Starting node")
	n.state = pbs.NodeState_ALIVE
	n.checkConnection(ctx)

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

			// The workers get 10 seconds to finish up.
			n.workers.Wait(time.Second * 10)
			return
		case <-ticker.C:
			n.sync(ctx)
			n.checkIdleTimer()
		}
	}
}

func (n *Node) checkConnection(ctx context.Context) {
	_, err := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})

	if err != nil {
		log.Error("Couldn't contact server.", err)
	} else {
		log.Info("Successfully connected to server.")
	}
}

// sync syncs the node's state with the server. It reports task state changes,
// handles signals from the server (new task, cancel task, etc), reports resources, etc.
//
// TODO Sync should probably use a channel to sync data access.
//      Probably only a problem for test code, where Sync is called directly.
func (n *Node) sync(ctx context.Context) {
	r, gerr := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})

	if gerr != nil {
		log.Error("Couldn't get node state during sync.", gerr)
		return
	}

	// Start task workers. runSet will track task IDs
	// to ensure there's only one runner per ID, so it's ok
	// to call this multiple times with the same task ID.
	for _, id := range r.TaskIds {
		if n.workers.Add(id) {
			go func(id string) {
				r := n.newWorker(n.workerConf, id)
				r.Run(ctx)
				n.workers.Remove(id)
			}(id)
		}
	}

	// Node data has been updated. Send back to server for database update.
	res := detectResources(n.conf)
	r.Resources = &pbs.Resources{
		Cpus:   res.Cpus,
		RamGb:  res.RamGb,
		DiskGb: res.DiskGb,
	}
	r.State = n.state

	// Merge metadata
	if r.Metadata == nil {
		r.Metadata = map[string]string{}
	}
	for k, v := range n.conf.Metadata {
		r.Metadata[k] = v
	}

	_, err := n.client.UpdateNode(context.Background(), r)
	if err != nil {
		log.Error("Couldn't save node update. Recovering.", err)
	}
}

// Check if the worker pool is idle. If so, start the timeout timer.
func (n *Node) checkIdleTimer() {
	// The pool is idle if there are no task runners.
	// The pool should not time out if it's not alive (e.g. if it's initializing)
	idle := n.workers.Count() == 0 && n.state == pbs.NodeState_ALIVE
	if idle {
		n.timeout.Start()
	} else {
		n.timeout.Stop()
	}
}
