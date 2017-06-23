package manual

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
)

var log = logger.NewSubLogger("manual")

// Plugin provides the manual scheduler backend plugin.
var Plugin = &scheduler.BackendPlugin{
	Name:   "manual",
	Create: NewBackend,
}

// NewBackend returns a new Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {

	// Create a client for talking to the funnel scheduler
	client, err := scheduler.NewClient(conf.Worker)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	return scheduler.Backend(&Backend{conf, client}), nil
}

// Backend represents the manual backend.
type Backend struct {
	conf   config.Config
	client scheduler.Client
}

// Schedule schedules a task on a manual worker instance.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running manual scheduler")

	offers := []*scheduler.Offer{}

	for _, w := range s.getWorkers() {
		// Filter out workers that don't match the task request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !scheduler.Match(w, j, scheduler.DefaultPredicates) {
			continue
		}

		sc := scheduler.DefaultScores(w, j)
		offer := scheduler.NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching workers were found.
	if len(offers) == 0 {
		return nil
	}

	scheduler.SortByAverageScore(offers)
	return offers[0]
}

func (s *Backend) getWorkers() []*pbf.Worker {

	// Get the workers from the funnel server
	workers := []*pbf.Worker{}
	req := &pbf.ListWorkersRequest{}
	resp, err := s.client.ListWorkers(context.Background(), req)

	// If there's an error, return an empty list
	if err != nil {
		log.Error("Failed ListWorkers request. Recovering.", err)
		return workers
	}

	return resp.Workers
}
