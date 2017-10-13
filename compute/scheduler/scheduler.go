package scheduler

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"time"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	QueueTask(*tes.Task) error
	ReadQueue(int) []*tes.Task
	ListNodes(context.Context, *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error)
	PutNode(context.Context, *pbs.Node) (*pbs.PutNodeResponse, error)
	DeleteNode(context.Context, *pbs.Node) error
	Write(ev *events.Event) error
}

// Scheduler handles scheduling tasks to nodes and support many backends.
type Scheduler struct {
	Log     *logger.Logger
	DB      Database
	Conf    config.Scheduler
	Backend Backend
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
			err = s.Scale(ctx)
			if err != nil {
				return fmt.Errorf("scale error: %s", err)
			}
		}
	}
}

// CheckNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (s *Scheduler) CheckNodes() error {
	ctx := context.Background()
	resp, err := s.DB.ListNodes(ctx, &pbs.ListNodesRequest{})

	if err != nil {
		return err
	}

	updated := UpdateNodeState(resp.Nodes, s.Conf)

	for _, node := range updated {
		var err error

		if node.State == pbs.NodeState_GONE {
			err = s.DB.DeleteNode(ctx, node)
		} else {
			_, err = s.DB.PutNode(ctx, node)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Schedule does a scheduling iteration. It checks the health of nodes
// in the database, gets a chunk of tasks from the queue (configurable by config.Scheduler.ScheduleChunk),
// and calls the given scheduler backend. If the backend returns a valid offer, the
// task is assigned to the offered node.
func (s *Scheduler) Schedule(ctx context.Context) error {
	err := s.CheckNodes()
	if err != nil {
		s.Log.Error("Error checking nodes", err)
	}

	for _, task := range s.DB.ReadQueue(s.Conf.ScheduleChunk) {
		offer := s.Backend.GetOffer(task)
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
			_, err = s.DB.PutNode(ctx, offer.Node)
			if err != nil {
				s.Log.Error("Error in AssignTask", err)
				continue
			}

			err = s.DB.Write(events.NewState(task.Id, 0, tes.State_INITIALIZING))
			if err != nil {
				s.Log.Error("Error marking task as initializing", err)
			}
		} else {
			s.Log.Debug("Scheduling failed for task", "taskID", task.Id)
		}
	}
	return nil
}

// Scale implements some common logic for allowing scheduler backends
// to poll the database, looking for nodes that need to be started
// and shutdown.
func (s *Scheduler) Scale(ctx context.Context) error {

	b, isScaler := s.Backend.(Scaler)
	// If the scheduler backend doesn't implement the Scaler interface,
	// stop here.
	if !isScaler {
		return nil
	}

	resp, err := s.DB.ListNodes(ctx, &pbs.ListNodesRequest{})
	if err != nil {
		s.Log.Error("Failed ListNodes request. Recovering.", err)
		return nil
	}

	for _, n := range resp.Nodes {

		if !b.ShouldStartNode(n) {
			continue
		}

		serr := b.StartNode(n)
		if serr != nil {
			s.Log.Error("Error starting node", serr)
			continue
		}

		// TODO should the Scaler instance handle this? Is it possible
		//      that Initializing is the wrong state in some cases?
		n.State = pbs.NodeState_INITIALIZING
		_, err := s.DB.PutNode(ctx, n)

		if err != nil {
			// TODO an error here would likely result in multiple nodes
			//      being started unintentionally. Not sure what the best
			//      solution is. Possibly a list of failed nodes.
			//
			//      If the scheduler and database API live on the same server,
			//      this *should* be very unlikely.
			s.Log.Error("Error marking node as initializing", err)
		}
	}
	return nil
}
