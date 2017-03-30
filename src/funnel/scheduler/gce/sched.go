package gce

// TODO
// - resource tracking via GCP APIs
// - provisioning limits, e.g. don't create more than 100 VMs, or
//   maybe use N VCPUs max, across all VMs
// - act on failed machines?
// - know how to shutdown machines

import (
	"context"
	"fmt"
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	sched "funnel/scheduler"
)

var log = logger.New("gce")

// NewScheduler returns a new Google Cloud Engine Scheduler instance.
func NewScheduler(conf config.Config) (sched.Scheduler, error) {
	// TODO need GCE scheduler config validation. If zone is missing, nothing works.

	// Create a client for talking to the funnel scheduler
	client, err := sched.NewClient(conf.Worker)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	// Create a client for talking to the GCE API
	gce, gerr := newClientFromConfig(conf)
	if gerr != nil {
		log.Error("Can't connect GCE client", gerr)
		return nil, gerr
	}

	s := &gceScheduler{
		conf:   conf,
		client: client,
		gce:    gce,
	}

	return s, nil
}

type gceScheduler struct {
	conf   config.Config
	client sched.Client
	gce    Client
}

// Schedule schedules a job on a Google Cloud VM worker instance.
func (s *gceScheduler) Schedule(j *tes.Job) *sched.Offer {
	log.Debug("Running GCE scheduler")

	offers := []*sched.Offer{}
	predicates := append(sched.DefaultPredicates, sched.WorkerHasTag("gce"))

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
		sc = sc.Weighted(s.conf.Schedulers.GCE.Weights)

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

// getWorkers returns a list of all GCE workers and appends a set of
// uninitialized workers, which the scheduler can use to create new worker VMs.
func (s *gceScheduler) getWorkers() []*pbf.Worker {

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

	// Include unprovisioned (template) workers.
	// This is how the scheduler can schedule jobs to workers that
	// haven't been started yet.
	for _, t := range s.gce.Templates() {
		t.Id = sched.GenWorkerID("gce")
		workers = append(workers, &t)
	}

	return workers
}

// ShouldStartWorker tells the scaler loop which workers
// belong to this scheduler backend, basically.
func (s *gceScheduler) ShouldStartWorker(w *pbf.Worker) bool {
	// Only start works that are uninitialized and have a gce template.
	tpl, ok := w.Metadata["gce-template"]
	return ok && tpl != "" && w.State == pbf.WorkerState_Uninitialized
}

// StartWorker calls out to GCE APIs to start a new worker instance.
func (s *gceScheduler) StartWorker(w *pbf.Worker) error {

	// Write the funnel worker config yaml to a string
	c := s.conf.Worker
	c.ID = w.Id
	c.Timeout = -1

	// Get the template ID from the worker metadata
	template, ok := w.Metadata["gce-template"]
	if !ok || template == "" {
		return fmt.Errorf("Could not get GCE template ID from metadata")
	}

	return s.gce.StartWorker(template, c)
}
