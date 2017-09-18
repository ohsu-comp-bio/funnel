package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// WorkerFactory is a function which creates a new worker instance.
type WorkerFactory func(c config.Worker, taskID string) (worker.Worker, error)

// NoopWorkerFactory returns a new NoopWorker.
func NoopWorkerFactory(c config.Worker, taskID string) (worker.Worker, error) {
	return worker.NoopWorker{}, nil
}
