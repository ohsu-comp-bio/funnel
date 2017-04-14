package mocks

import (
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"funnel/scheduler"
	"funnel/worker"
)

// NewNoopWorker returns a new NoopWorker instance.
func NewNoopWorker(conf config.Config) *worker.Worker {
	conf.Worker.ID = "noop-worker"
	w, _ := worker.NewWorker(conf.Worker)
	// Stub the job runner so it's a no-op runner
	// i.e. ensure docker run, file copying, etc. doesn't actually happen
	w.JobRunner = worker.NoopJobRunner
	return w
}

// NoopBackend is a scheduler backend plugin useful for testing.
// It exposes the Worker instance, so tests can check state directly
// in the worker. The worker is configured with a NoopJobRunner,
// which avoids interaction with the storage and containers
// (i.e. no file downloads/uploads nor docker calls)
type NoopBackend struct {
	Worker *worker.Worker
	conf   config.Config
}

// Schedule schedules a job to the noop worker. There is only
// one worker and jobs are always scheduled to that worker without
// and logic or filtering (just dead simple).
func (s *NoopBackend) Schedule(j *tes.Job) *scheduler.Offer {
	w := &pbf.Worker{
		Id:    "noop-worker",
		State: pbf.WorkerState_Alive,
		Jobs:  map[string]*pbf.JobWrapper{},
	}
	return scheduler.NewOffer(w, j, scheduler.Scores{})
}
