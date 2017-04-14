package scheduler

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
)

// Backend is responsible for scheduling a task. It has a single method which
// is responsible for taking a Task and returning an Offer, or nil if there is
// no worker matching the task request. An Offer includes the ID of the offered
// worker.
//
// Offers include scores which describe how well the task fits the worker.
// Scores may describe a wide variety of metrics: resource usage, packing,
// startup time, cost, etc. Scores and weights are used to control the behavior
// of schedulers, and to combine offers from multiple schedulers.
type Backend interface {
	Schedule(*tes.Task) *Offer
}

// Scaler represents a service that can start worker instances, for example
// the Google Cloud Scheduler backend.
type Scaler interface {
	// StartWorker is where the work is done to start a worker instance,
	// for example, calling out to Google Cloud APIs.
	StartWorker(*pbf.Worker) error
	// ShouldStartWorker allows scalers to filter out workers they are interested in.
	// If "true" is returned, Scaler.StartWorker() will be called with this worker.
	ShouldStartWorker(*pbf.Worker) bool
}

// Offer describes a worker offered by a scheduler for a task.
// The Scores describe how well the task fits this worker,
// which could be used by other a scheduler to pick the best offer.
type Offer struct {
	TaskID string
	Worker *pbf.Worker
	Scores Scores
}

// NewOffer returns a new Offer instance.
func NewOffer(w *pbf.Worker, t *tes.Task, s Scores) *Offer {
	return &Offer{
		TaskID: t.Id,
		Worker: w,
		Scores: s,
	}
}

// BackendPlugin is provided by backends when they register with Scheduler,
// which allows to the scheduler to create a backend instance by name.
type BackendPlugin struct {
	Name     string
	Create   func(config.Config) (Backend, error)
	instance Backend
}
