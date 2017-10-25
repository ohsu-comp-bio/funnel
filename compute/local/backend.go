package local

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.Config, log *logger.Logger, fac scheduler.WorkerFactory) *Backend {
	return &Backend{conf, log, fac}
}

// Backend represents the local backend.
type Backend struct {
	conf      config.Config
	log       *logger.Logger
	newWorker scheduler.WorkerFactory
}

// Submit submits a task. For the Local backend this results in the task
// running immediately.
func (b *Backend) Submit(task *tes.Task) error {
	w, err := b.newWorker(b.conf.Worker, task.Id, b.log)
	if err != nil {
		return err
	}
	go w.Run(context.Background())
	return nil
}
