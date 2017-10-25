package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// NewNode returns a new Node instance
func NewNode(conf config.Config, log *logger.Logger, factory WorkerFactory) (*Node, error) {
	log = log.WithFields("nodeID", conf.Scheduler.Node.ID)

	cli, err := NewClient(conf.Scheduler)
	if err != nil {
		return nil, err
	}

	err = util.EnsureDir(conf.Scheduler.Node.WorkDir)
	if err != nil {
		return nil, err
	}

	// Detect available resources at startup
	res, derr := detectResources(conf.Scheduler.Node)
	if derr != nil {
		log.Error("error detecting resources", "error", derr)
	}

	timeout := util.NewIdleTimeout(conf.Scheduler.Node.Timeout)
	state := pbs.NodeState_UNINITIALIZED

	workerConf := conf.Worker
	workerConf.WorkDir = conf.Scheduler.Node.WorkDir

	return &Node{
		conf:       conf.Scheduler.Node,
		workerConf: workerConf,
		client:     cli,
		log:        log,
		resources:  res,
		newWorker:  factory,
		workers:    newRunSet(),
		timeout:    timeout,
		state:      state,
	}, nil
}

// NewNoopNode returns a new node that doesn't have any side effects
// (e.g. storage access, docker calls, etc.) which is useful for testing.
func NewNoopNode(conf config.Config, log *logger.Logger) (*Node, error) {
	return NewNode(conf, log, NoopWorkerFactory)
}

// Node is a structure used for tracking available resources on a compute resource.
type Node struct {
	conf       config.Node
	workerConf config.Worker
	client     Client
	log        *logger.Logger
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

	// If its a 404 error create a new node
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		n.log.Error("Couldn't contact server.", err)
	} else {
		n.log.Info("Successfully connected to server.")
	}
}

// sync syncs the node's state with the server. It reports task state changes,
// handles signals from the server (new task, cancel task, etc), reports resources, etc.
//
// TODO Sync should probably use a channel to sync data access.
//      Probably only a problem for test code, where Sync is called directly.
func (n *Node) sync(ctx context.Context) {
	var r *pbs.Node
	var err error

	r, err = n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})
	if err != nil {
		// If its a 404 error create a new node
		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			n.log.Error("Couldn't get node state during sync.", err)
			return
		}
		n.log.Info("Starting initial node sync")
		r = &pbs.Node{Id: n.conf.ID}
	}

	// Start task workers. runSet will track task IDs
	// to ensure there's only one worker per ID, so it's ok
	// to call this multiple times with the same task ID.
	for _, id := range r.TaskIds {
		if n.workers.Add(id) {
			go n.runTask(ctx, id)
		}
	}

	// Node data has been updated. Send back to server for database update.
	var derr error
	n.resources, derr = detectResources(n.conf)
	if derr != nil {
		n.log.Error("error detecting resources", "error", derr)
	}

	// Merge metadata
	meta := map[string]string{}
	for k, v := range n.conf.Metadata {
		meta[k] = v
	}
	for k, v := range r.GetMetadata() {
		meta[k] = v
	}

	_, err = n.client.PutNode(context.Background(), &pbs.Node{
		Id:        n.conf.ID,
		Resources: &n.resources,
		State:     n.state,
		Version:   r.GetVersion(),
		Metadata:  meta,
		TaskIds:   r.TaskIds,
	})
	if err != nil {
		n.log.Error("Couldn't save node update. Recovering.", err)
	}
}

func (n *Node) runTask(ctx context.Context, id string) {
	log := n.log.WithFields("ns", "worker", "taskID", id)
	// TODO handle error
	r, _ := n.newWorker(n.workerConf, id, log)
	r.Run(ctx)
	defer n.workers.Remove(id)

	// task cannot fully complete until it has successfully removed the
	// assigned ID from the node database. this helps prevent tasks from
	// running multiple times.
	ticker := time.NewTicker(n.conf.UpdateRate)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r, err := n.client.GetNode(ctx, &pbs.GetNodeRequest{Id: n.conf.ID})
			if err != nil {
				log.Error("couldn't get node state during task sync.", err)
				// break out of "select", but not "for".
				// i.e. try again later
				break
			}

			// Find the finished task ID in the node's assigned IDs and remove it.
			var ids []string
			for _, tid := range r.TaskIds {
				if tid != id {
					ids = append(ids, tid)
				}
			}
			r.TaskIds = ids

			_, err = n.client.PutNode(ctx, r)
			if err != nil {
				log.Error("couldn't save node update during task sync.", err)
				// break out of "select", but not "for".
				// i.e. try again later
				break
			}
			// Update was successful, return.
			return
		}
	}
}

// Check if the node is idle. If so, start the timeout timer.
func (n *Node) checkIdleTimer() {
	// The pool is idle if there are no task workers.
	// The pool should not time out if it's not alive (e.g. if it's initializing)
	idle := n.workers.Count() == 0 && n.state == pbs.NodeState_ALIVE
	if idle {
		n.timeout.Start()
	} else {
		n.timeout.Stop()
	}
}
