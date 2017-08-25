package node

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// NoopWorkerFactory returns a new NoopWorker.
func NoopWorkerFactory(c config.Worker, taskID string) worker.Worker {
	return worker.NoopWorker{}
}
