// Package scheduler contains Funnel's builtin compute scheduler and node.
package scheduler

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
)

// TaskQueue describes the interface the scheduler uses to find tasks that need scheduling.
type TaskQueue interface {
	ReadQueue(count int) []*tes.Task
}

// Scheduler handles scheduling tasks to nodes and support many backends.
type Scheduler struct {
	Conf  config.Scheduler
	Log   *logger.Logger
	Queue TaskQueue
	Event events.Writer

	handles map[string]*nodeHandle
	mtx     sync.Mutex
}

func NewScheduler(conf config.Scheduler, log *logger.Logger, event events.Writer) *Scheduler {
	return &Scheduler{
		Conf:    conf,
		Log:     log,
		Event:   event,
		handles: map[string]*nodeHandle{},
	}
}

type nodeHandle struct {
	node *Node
	send func(*tes.Task) error
}

func (s *Scheduler) NodeChat(stream SchedulerService_NodeChatServer) error {

	// Ensure the node is marked as dead immediately when the connection is dropped.
	var node *Node
	defer func() {
		if node != nil {
			node.State = NodeState_DEAD
		}
	}()

	// Constantly receive updates from the node.
	for {
		var err error
		node, err = stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Update the node record
		s.mtx.Lock()
		s.handles[node.Id] = &nodeHandle{
			node: node,
			send: stream.Send,
		}
		s.mtx.Unlock()
	}

	return nil
}

func (s *Scheduler) ListNodes(ctx context.Context, req *ListNodesRequest) (*ListNodesResponse, error) {
	return &ListNodesResponse{Nodes: s.nodes()}, nil
}

func (s *Scheduler) GetNode(ctx context.Context, req *GetNodeRequest) (*Node, error) {
	h, ok := s.handles[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.Id)
	}
	return h.node, nil
}

func (s *Scheduler) CreateTask(ctx context.Context, task *tes.Task) error {
	// TODO write to the database
	//      write to a channel for immediate scheduling
	s.schedule(task)
	return nil
}

func (s *Scheduler) CancelTask(ctx context.Context, taskID string) error {
	// CancelTask is a noop in the manual scheduler.
	// Nodes/Workers will pick up the canceled state via polling.
	return nil
}

// Run starts the scheduling loop. This blocks.
//
// The scheduler will take a chunk of tasks from the queue,
// request the the configured backend schedule them, and
// act on offers made by the backend.
func (s *Scheduler) Run(ctx context.Context) error {
	for range util.Ticker(ctx, time.Duration(s.Conf.ScheduleRate)) {
		s.checkNodes(ctx)

		/* TODO periodically schedule all queued tasks
		   err := s.schedule(ctx)
		   if err != nil {
		     s.Log.Error("error scheduling tasks: %s", err)
		   }
		*/
	}

	return nil
}

// checkNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (s *Scheduler) checkNodes(ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for id, h := range s.handles {
		s.updateNodeState(h.node)
		if h.node.State == NodeState_GONE {
			/* TODO get tasks assigned to node
			for _, tid := range node.TaskIds {
				s.Event.WriteEvent(ctx, events.NewState(tid, tes.State_SYSTEM_ERROR))
				s.Event.WriteEvent(ctx, events.NewSystemLog(tid, 0, 0, "info",
					"Cleaning up Task assigned to dead/gone node", map[string]string{
						"nodeID": node.Id,
					}))
			}
			*/
			delete(s.handles, id)
		}
	}
}

func (s *Scheduler) nodes() []*Node {
	var nodes []*Node
	for _, h := range s.handles {
		nodes = append(nodes, h.node)
	}
	return nodes
}

func (s *Scheduler) schedule(task *tes.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	offer := DefaultScheduleAlgorithm(task, s.nodes(), nil)
	if offer == nil {
		return fmt.Errorf("no offer for task %s", task.Id)
	}
	h := s.handles[offer.Node.Id]
	return h.send(task)

	/* TODO move to node
	s.Event.WriteEvent(ctx, events.NewMetadata(task.Id, 0, map[string]string{
	  "nodeID", offer.Node.Id,
	}))
	*/
}

// updateNodeState checks whether a node is dead/gone based on the last
// time it pinged.
func (s *Scheduler) updateNodeState(node *Node) {

	if node.State == NodeState_GONE {
		return
	}

	if node.LastPing == 0 {
		// This shouldn't be happening, because nodes should be
		// created with LastPing, but give it the benefit of the doubt
		// and leave it alone.
		return
	}

	lastPing := time.Unix(0, node.LastPing)
	d := time.Since(lastPing)

	if node.State == NodeState_UNINITIALIZED || node.State == NodeState_INITIALIZING {

		// The node is initializing, which has a more liberal timeout.
		if d > time.Duration(s.Conf.NodeInitTimeout) {
			// Looks like the node failed to initialize. Mark it dead
			node.State = NodeState_DEAD
		}

	} else if node.State == NodeState_DEAD && d > time.Duration(s.Conf.NodeDeadTimeout) {
		// The node has been dead for long enough.
		node.State = NodeState_GONE

	} else if d > time.Duration(s.Conf.NodePingTimeout) {
		// The node hasn't pinged in awhile, mark it dead.
		node.State = NodeState_DEAD
	}
}
