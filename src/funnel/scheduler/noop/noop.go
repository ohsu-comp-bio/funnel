package noop

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"funnel/scheduler"
	"funnel/worker"
)

// Config updates the config with noop backend values
// and returns a new config instance
func Config(conf config.Config) config.Config {
	conf.Scheduler = "noop"
	return conf
}

// NewPlugin returns a new scheduler backend plugin configured
// to use the given worker.
func NewPlugin(w *worker.Worker) *scheduler.BackendPlugin {
	return &scheduler.BackendPlugin{
		Name: "noop",
		Create: func(conf config.Config) (scheduler.Backend, error) {
			return &Backend{w, conf}, nil
		},
	}
}

// NewWorker returns a new Worker instance.
func NewWorker(conf config.Config) *worker.Worker {
	conf.Worker.ID = "noop-worker"
	w, _ := worker.NewWorker(conf.Worker)
	// Stub the task runner so it's a no-op runner
	// i.e. ensure docker run, file copying, etc. doesn't actually happen
	w.TaskRunner = worker.NoopTaskRunner
	return w
}

// Backend is a scheduler backend plugin useful for testing.
// It exposes the Worker instance, so tests can check state directly
// in the worker. The worker is configured with a NoopTaskRunner,
// which avoids interaction with the storage and containers
// (i.e. no file downloads/uploads nor docker calls)
type Backend struct {
	Worker *worker.Worker
	conf   config.Config
}

// Schedule schedules a task to the noop worker. There is only
// one worker and tasks are always scheduled to that worker without
// and logic or filtering (just dead simple).
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	w := &pbf.Worker{
		Id:    "noop-worker",
		State: pbf.WorkerState_Alive,
		Tasks: map[string]*pbf.TaskWrapper{},
	}
	return scheduler.NewOffer(w, j, scheduler.Scores{})
}
