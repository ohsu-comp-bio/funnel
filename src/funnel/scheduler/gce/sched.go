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
	"funnel/scheduler"
)

var log = logger.New("gce")

// Plugin provides the Google Cloud Compute scheduler backend plugin.
var Plugin = &scheduler.BackendPlugin{
	Name:   "gce",
	Create: NewBackend,
}

// NewBackend returns a new Google Cloud Engine Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	// TODO need GCE scheduler config validation. If zone is missing, nothing works.

	// Create a client for talking to the funnel scheduler
	client, err := scheduler.NewClient(conf.Worker)
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

	s := &Backend{
		conf:   conf,
		client: client,
		gce:    gce,
	}

	return scheduler.Backend(s), nil
}

// Backend represents the GCE backend, which provides
// and interface for both scheduling and scaling.
type Backend struct {
	conf   config.Config
	client scheduler.Client
	gce    Client
}

// Schedule schedules a task on a Google Cloud VM worker instance.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running GCE scheduler")

	offers := []*scheduler.Offer{}
	predicates := append(scheduler.DefaultPredicates, scheduler.WorkerHasTag("gce"))

	for _, w := range s.getWorkers() {
		// Filter out workers that don't match the task request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !scheduler.Match(w, j, predicates) {
			continue
		}

		sc := scheduler.DefaultScores(w, j)
		/*
			    TODO?
			    if w.State == pbf.WorkerState_Alive {
					  sc["startup time"] = 1.0
			    }
		*/
		sc = sc.Weighted(s.conf.Backends.GCE.Weights)

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

// getWorkers returns a list of all GCE workers and appends a set of
// uninitialized workers, which the scheduler can use to create new worker VMs.
func (s *Backend) getWorkers() []*pbf.Worker {

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
	// This is how the scheduler can schedule tasks to workers that
	// haven't been started yet.
	for _, t := range s.gce.Templates() {
		t.Id = scheduler.GenWorkerID("gce")
		workers = append(workers, &t)
	}

	return workers
}

// ShouldStartWorker tells the scaler loop which workers
// belong to this scheduler backend, basically.
func (s *Backend) ShouldStartWorker(w *pbf.Worker) bool {
	// Only start works that are uninitialized and have a gce template.
	tpl, ok := w.Metadata["gce-template"]
	return ok && tpl != "" && w.State == pbf.WorkerState_Uninitialized
}

// StartWorker calls out to GCE APIs to start a new worker instance.
func (s *Backend) StartWorker(w *pbf.Worker) error {

	// Write the funnel worker config yaml to a string
	wc := s.conf
	wc.Worker.ID = w.Id
	wc.Worker.Timeout = -1

	// Get the template ID from the worker metadata
	template, ok := w.Metadata["gce-template"]
	if !ok || template == "" {
		return fmt.Errorf("Could not get GCE template ID from metadata")
	}

	return s.gce.StartWorker(template, wc)
}
