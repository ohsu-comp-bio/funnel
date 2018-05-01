// Package scheduler contains Funnel's builtin compute scheduler and node.
package builtin

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// Scheduler handles scheduling tasks to nodes and support many backends.
type Scheduler struct {
	conf  config.Scheduler
	log   *logger.Logger
	event events.Writer

	db      *badger.DB
	handles map[string]*nodeHandle
	mtx     sync.Mutex
	trigger *trigger
}

type nodeHandle struct {
	node *Node
	send func(*Control) error
}

// NewScheduler creates a new scheduler instance, which creates the database.
func NewScheduler(conf config.Scheduler, log *logger.Logger, event events.Writer) (*Scheduler, error) {
	err := fsutil.EnsureDir(conf.DBPath)
	if err != nil {
		return nil, fmt.Errorf("creating database directory: %s", err)
	}

	opts := badger.DefaultOptions
	opts.Dir = conf.DBPath
	opts.ValueDir = conf.DBPath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("opening database: %s", err)
	}

	return &Scheduler{
		db:      db,
		conf:    conf,
		log:     log,
		event:   event,
		handles: map[string]*nodeHandle{},
		trigger: newTrigger(),
	}, nil
}

// NodeChat handles online, bidirectional, streaming communication
// between a node and the scheduler.
func (s *Scheduler) NodeChat(stream SchedulerService_NodeChatServer) error {
	if s == nil {
		return fmt.Errorf("scheduler is nil")
	}
	// gRPC starts a separate NodeChat process for each node connection.

	// Ensure the node is marked as dead immediately when the connection is dropped.
	var node *Node
	defer func() {
		if node != nil && node.State != NodeState_GONE {
			node.State = NodeState_DEAD
		}
	}()

	// Constantly receive updates from the node.
	for {
		n, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		node = n

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

// ListNodes is an API endpoint which returns a list of nodes.
func (s *Scheduler) ListNodes(ctx context.Context, req *ListNodesRequest) (*ListNodesResponse, error) {
	return &ListNodesResponse{Nodes: s.nodes()}, nil
}

// GetNode is an API endpoint which returns a single node.
func (s *Scheduler) GetNode(ctx context.Context, req *GetNodeRequest) (*Node, error) {
	h, ok := s.handles[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.Id)
	}
	return h.node, nil
}

// DrainNode is an API endpoint which sets a node's state to DRAIN,
// which causes the node to stop accepting jobs.
func (s *Scheduler) DrainNode(ctx context.Context, req *DrainNodeRequest) (*DrainNodeResponse, error) {
	h, ok := s.handles[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.Id)
	}
	err := h.send(&Control{Type: ControlType_DRAIN_NODE})
	return &DrainNodeResponse{}, err
}

// Submit accepts a new task for scheduling.
func (s *Scheduler) Submit(ctx context.Context, task *tes.Task) error {
	// Save task to queue
	err := s.db.Update(func(txn *badger.Txn) error {
		b, err := proto.Marshal(task)
		if err != nil {
			return fmt.Errorf("marshaling task: %s", err)
		}
		return txn.Set([]byte(task.Id), b)
	})
	if err != nil {
		s.log.Error("error saving task")
		return fmt.Errorf("saving task to scheduler queue: %s", err)
	}

	s.trigger.trigger()
	return nil
}

// Cancel is a noop in the builtin scheduler.
// Nodes/Workers will pick up the canceled state via polling.
func (s *Scheduler) Cancel(ctx context.Context, taskID string) error {
	return nil
}

// Run starts the scheduling loop. This blocks.
//
// The scheduler will take a chunk of tasks from the queue,
// request the the configured backend schedule them, and
// act on offers made by the backend.
func (s *Scheduler) Run(ctx context.Context) {
	go func() {
		for range s.trigger.ch {
			// TODO will contention for the mutex cause scheduling to be delayed?
			s.checkNodes(ctx)
			s.scheduleMany()
		}
	}()

	for range util.Ticker(ctx, time.Duration(100*time.Millisecond)) {
		s.trigger.trigger()
	}
}

// scheduleMany reads queued tasks from the database and schedules them.
func (s *Scheduler) scheduleMany() {
	s.db.Update(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {

			task := &tes.Task{}
			v, err := it.Item().Value()
			if err != nil {
				s.log.Error("error getting task value", err)
				continue
			}

			err = proto.Unmarshal(v, task)
			if err != nil {
				s.log.Error("error unmarshaling task", err)
				continue
			}

			err = s.scheduleOne(task)
			if err != nil {
				s.log.Debug("error scheduling task", err)
				continue
			}

			err = txn.Delete(it.Item().Key())
			if err != nil {
				s.log.Error("error deleting task", err)
				continue
			}
		}
		return nil
	})
}

// scheduleOne schedules the given task.
func (s *Scheduler) scheduleOne(task *tes.Task) error {

	offer := DefaultScheduleAlgorithm(task, s.nodes(), nil)
	if offer == nil {
		return &noOfferError{task.Id}
	}

	err := s.assignTask(task, offer.Node.Id)
	if err != nil {
		return fmt.Errorf("assigning task to node: %s", err)
	}
	return nil
}

func (s *Scheduler) assignTask(task *tes.Task, nodeID string) error {
	h, ok := s.handles[nodeID]
	if !ok {
		return fmt.Errorf("no such node: %s", nodeID)
	}

	// TODO this doesn't immediately adjust the available resources,
	//      so scheduling happens fast, assignments will be wrong.
	err := h.send(&Control{
		Type: ControlType_CREATE_TASK,
		Task: task,
	})
	if err != nil {
		return fmt.Errorf("sending task to node: %s", err)
	}

	s.event.WriteEvent(context.Background(), events.NewMetadata(task.Id, 0, map[string]string{
		"nodeID": nodeID,
	}))
	return nil
}

// checkNodes checks for dead/gone nodes.
func (s *Scheduler) checkNodes(ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for id, h := range s.handles {
		s.updateNodeState(h.node)

		// Clean up tasks assigned to a failed node.
		if h.node.State == NodeState_GONE {
			for _, tid := range h.node.TaskIds {
				s.event.WriteEvent(ctx, events.NewState(tid, tes.State_SYSTEM_ERROR))
				s.event.WriteEvent(ctx, events.NewSystemLog(tid, 0, 0, "info",
					"Cleaning up Task assigned to dead/gone node", map[string]string{
						"nodeID": h.node.Id,
					}))
			}
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

	if node.State == NodeState_DEAD && d > time.Duration(s.conf.NodeDeadTimeout) {
		// The node has been dead for long enough.
		node.State = NodeState_GONE

	} else if d > time.Duration(s.conf.NodePingTimeout) {
		// The node hasn't pinged in awhile, mark it dead.
		node.State = NodeState_DEAD
	}
}

// trigger ensures that only one scheduling iteration
// is in-flight or pending at once.
type trigger struct {
	ch chan struct{}
}

func newTrigger() *trigger {
	// buffer size of 1 means that one scheduling iteration may be pending.
	return &trigger{ch: make(chan struct{}, 1)}
}
func (t trigger) trigger() {
	select {
	case t.ch <- struct{}{}:
	default:
		// If the channel can't be written to, do nothing.
		// This means there's already one iteration running and one pending.
	}
}

type noOfferError struct {
	taskID string
}

func (e *noOfferError) Error() string {
	return fmt.Sprintf("no offer for task %s", e.taskID)
}
