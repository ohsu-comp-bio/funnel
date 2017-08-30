package openstack

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
)

// Name of the scheduler backend
const Name = "openstack"

var log = logger.Sub(Name)

// NewBackend returns a new Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {

	// Create a client for talking to the funnel node
	client, err := scheduler.NewClient(conf.Scheduler)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	return &Backend{conf, client}, nil
}

// Backend represents the OpenStack backend.
type Backend struct {
	conf   config.Config
	client scheduler.Client
}

// Schedule schedules a task on a OpenStack VM node instance.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running OpenStack scheduler")

	offers := []*scheduler.Offer{}
	predicates := append(scheduler.DefaultPredicates, scheduler.NodeHasTag("openstack"))

	for _, w := range s.getNodes() {
		// Filter out nodes that don't match the task request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !scheduler.Match(w, j, predicates) {
			continue
		}

		sc := scheduler.DefaultScores(w, j)
		/*
				    TODO?
				    if w.State == pbs.NodeState_Alive {
						  sc["startup time"] = 1.0
				    }
			sc = sc.Weighted(s.conf.Backends.OpenStack.Weights)
		*/

		offer := scheduler.NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching nodes were found.
	if len(offers) == 0 {
		return nil
	}

	scheduler.SortByAverageScore(offers)
	return offers[0]
}

func (s *Backend) getNodes() []*pbs.Node {

	// Get the nodes from the funnel server
	nodes := []*pbs.Node{}
	req := &pbs.ListNodesRequest{}
	resp, err := s.client.ListNodes(context.Background(), req)

	// If there's an error, return an empty list
	if err != nil {
		log.Error("Failed ListNodes request. Recovering.", err)
		return nodes
	}

	nodes = resp.Nodes

	// TODO include unprovisioned node templates from config

	return nodes
}
