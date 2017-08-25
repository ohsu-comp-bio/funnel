package local

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/node"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"golang.org/x/net/context"
)

// Name of the scheduler backend.
const Name = "local"

var log = logger.Sub(Name)

// NewBackend returns a new Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	id := scheduler.GenNodeID("local")

	err := startNode(id, conf)
	if err != nil {
		return nil, err
	}

	client, err := node.NewClient(conf.Scheduler.Node)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	return &Backend{conf, client, id}, nil
}

// Backend represents the local backend.
type Backend struct {
	conf   config.Config
	client node.Client
	nodeID string
}

// Schedule schedules a task.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running local scheduler backend")
	weights := map[string]float32{}
	nodes := s.getNodes()
	return scheduler.DefaultScheduleAlgorithm(j, nodes, weights)
}

// getNodes gets a list of active, local nodes.
//
// This is a bit redundant in the local scheduler, because there is only
// ever one node, but it demonstrates the pattern of a scheduler backend,
// and give the scheduler a chance to get an updated node state.
func (s *Backend) getNodes() []*pbs.Node {
	nodes := []*pbs.Node{}
	resp, rerr := s.client.ListNodes(context.Background(), &pbs.ListNodesRequest{})

	if rerr != nil {
		log.Error("Error getting nodes. Recovering.", rerr)
		return nodes
	}

	for _, n := range resp.Nodes {
		if n.Id != s.nodeID || n.State != pbs.NodeState_ALIVE {
			// Ignore nodes that aren't alive
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func startNode(id string, conf config.Config) error {
	conf.Scheduler.Node.ID = id
	conf.Scheduler.Node.Timeout = -1

	log.Debug("Starting local node")

	n, err := node.NewNode(conf)
	if err != nil {
		log.Error("Can't create node", err)
		return err
	}
	go n.Start(context.Background())
	return nil
}
