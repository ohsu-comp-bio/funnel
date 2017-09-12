package manual

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
)

var log = logger.Sub("manual")

// NewBackend returns a new Backend instance.
func NewBackend(conf config.Config) (*Backend, error) {
	// Create a client for talking to the funnel scheduler
	client, err := scheduler.NewClient(conf.Scheduler)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	return &Backend{conf, client}, nil
}

// Backend represents the manual backend.
type Backend struct {
	conf   config.Config
	client scheduler.Client
}

// Schedule schedules a task on a manual node instance.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running manual scheduler")

	offers := []*scheduler.Offer{}

	for _, n := range s.getNodes() {
		// Filter out nodes that don't match the task request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !scheduler.Match(n, j, scheduler.DefaultPredicates) {
			continue
		}

		sc := scheduler.DefaultScores(n, j)
		offer := scheduler.NewOffer(n, j, sc)
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

	return resp.Nodes
}
