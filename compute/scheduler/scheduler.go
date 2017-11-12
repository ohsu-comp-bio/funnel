package scheduler

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// Nodes defines the database interface the scheduler uses for
// accessing nodes.
type Nodes interface {
	ListNodes(context.Context, *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error)
	PutNode(context.Context, *pbs.Node) (*pbs.PutNodeResponse, error)
	DeleteNode(context.Context, *pbs.Node) error
}

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	QueueTask(*tes.Task) error
	ReadQueue(int) []*tes.Task
	WriteContext(context.Context, *events.Event) error
}

// Scheduler handles scheduling tasks to nodes and support many backends.
type Scheduler struct {
	Log   *logger.Logger
	DB    Database
	Nodes Nodes
	Conf  config.Scheduler
}

// Submit submits a task via gRPC call to the funnel scheduler backend
func (s *Scheduler) Submit(task *tes.Task) error {
	err := s.DB.QueueTask(task)
	if err != nil {
		return fmt.Errorf("Failed to submit task %s to the scheduler queue: %s", task.Id, err)
	}
	return nil
}

// Run starts the scheduling loop. This blocks.
//
// The scheduler will take a chunk of tasks from the queue,
// request the the configured backend schedule them, and
// act on offers made by the backend.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.Conf.ScheduleRate)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var err error
			err = s.Schedule(ctx)
			if err != nil {
				return fmt.Errorf("schedule error: %s", err)
			}
		}
	}
}

// CheckNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (s *Scheduler) CheckNodes() error {
	ctx := context.Background()
	resp, err := s.Nodes.ListNodes(ctx, &pbs.ListNodesRequest{})

	if err != nil {
		return err
	}

	updated := UpdateNodeState(resp.Nodes, s.Conf)

	for _, node := range updated {
		var err error

		if node.State == pbs.NodeState_GONE {
			err = s.Nodes.DeleteNode(ctx, node)
		} else {
			_, err = s.Nodes.PutNode(ctx, node)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Schedule does a scheduling iteration. It checks the health of nodes
// in the database, gets a chunk of tasks from the queue (configurable by config.ScheduleChunk),
// and calls the given scheduler backend. If the backend returns a valid offer, the
// task is assigned to the offered node.
func (s *Scheduler) Schedule(ctx context.Context) error {
	err := s.CheckNodes()
	if err != nil {
		s.Log.Error("Error checking nodes", err)
	}

	for _, task := range s.DB.ReadQueue(s.Conf.ScheduleChunk) {
		offer := s.GetOffer(task)
		if offer != nil {
			s.Log.Info("Assigning task to node",
				"taskID", task.Id,
				"nodeID", offer.Node.Id,
				"node", offer.Node,
			)

			// TODO this is important! write a test for this line.
			//      when a task is assigned, its state is immediately Initializing
			//      even before the node has received it.
			offer.Node.TaskIds = append(offer.Node.TaskIds, task.Id)
			_, err = s.Nodes.PutNode(ctx, offer.Node)
			if err != nil {
				s.Log.Error("Error in AssignTask",
					"error", err,
					"taskID", task.Id,
					"nodeID", offer.Node.Id,
				)
				continue
			}

			err = s.DB.WriteContext(ctx, events.NewState(task.Id, 0, tes.State_INITIALIZING))
			if err != nil {
				s.Log.Error("Error marking task as initializing",
					"error", err,
					"taskID", task.Id,
					"nodeID", offer.Node.Id,
				)
			}
		} else {
			s.Log.Debug("Scheduling failed for task", "taskID", task.Id)
		}
	}
	return nil
}

// GetOffer returns an offer based on available funnel nodes.
func (s *Scheduler) GetOffer(j *tes.Task) *Offer {
	offers := []*Offer{}

	// Get the nodes from the funnel server
	nodes := []*pbs.Node{}
	resp, err := s.Nodes.ListNodes(context.Background(), &pbs.ListNodesRequest{})
	if err == nil {
		nodes = resp.Nodes
	}

	for _, n := range nodes {
		// Only schedule tasks to nodes that are "ALIVE"
		if n.State != pbs.NodeState_ALIVE {
			continue
		}
		// Filter out nodes that don't match the task request.
		// Checks CPU, RAM, disk space, etc.
		if !Match(n, j, DefaultPredicates) {
			continue
		}

		sc := DefaultScores(n, j)
		offer := NewOffer(n, j, sc)
		offers = append(offers, offer)
	}

	// No matching nodes were found.
	if len(offers) == 0 {
		return nil
	}

	SortByAverageScore(offers)
	return offers[0]
}
