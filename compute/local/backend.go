package local

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/worker"
)

var log = logger.Sub("local")

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.Config) *Backend {
	return &Backend{conf}
}

// Backend represents the local backend.
type Backend struct {
	conf config.Config
}

// Submit submits a task. For the Local backend this results in the task
// running immediately.
func (b *Backend) Submit(task *tes.Task) error {
	log.Debug("Submitting to local", "taskID", task.Id)
	w, err := worker.NewDefaultWorker(b.conf.Worker)
	if err != nil {
		return err
	}
	go w.Run(context.Background(), task)
	return nil
}
