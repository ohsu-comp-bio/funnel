package openstack

import (
	"context"
	"funnel/config"
	tes "funnel/proto/tes"
	"funnel/logger"
	sched "funnel/scheduler"
	pbf "funnel/proto/funnel"
)

var log = logger.New("openstack")

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf config.Config) (sched.Scheduler, error) {

	// Create a client for talking to the funnel scheduler
	client, err := sched.NewClient(conf.Worker)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	return &scheduler{conf, client}, nil
}

type scheduler struct {
	conf   config.Config
	client sched.Client
}

// Schedule schedules a job on a OpenStack VM worker instance.
func (s *scheduler) Schedule(j *tes.Job) *sched.Offer {
	log.Debug("Running OpenStack scheduler")

	offers := []*sched.Offer{}
	predicates := append(sched.DefaultPredicates, sched.WorkerHasTag("openstack"))

	for _, w := range s.getWorkers() {
		// Filter out workers that don't match the job request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !sched.Match(w, j, predicates) {
			continue
		}

		sc := sched.DefaultScores(w, j)
		/*
			    TODO?
			    if w.State == pbf.WorkerState_Alive {
					  sc["startup time"] = 1.0
			    }
		*/
		sc = sc.Weighted(s.conf.Schedulers.OpenStack.Weights)

		offer := sched.NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching workers were found.
	if len(offers) == 0 {
		return nil
	}

	sched.SortByAverageScore(offers)
	return offers[0]
}

func (s *scheduler) getWorkers() []*pbf.Worker {

	// Get the workers from the funnel server
	workers := []*pbf.Worker{}
	req := &pbf.GetWorkersRequest{}
	resp, err := s.client.GetWorkers(context.Background(), req)

	// If there's an error, return an empty list
	if err != nil {
		log.Error("Failed GetWorkers request. Recovering.", err)
		return workers
	}

	workers = resp.Workers

	// TODO include unprovisioned worker templates from config

	return workers
}
