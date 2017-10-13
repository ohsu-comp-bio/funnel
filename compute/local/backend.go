package local

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.Config, log *logger.Logger) *Backend {
	return &Backend{conf, log}
}

// Backend represents the local backend.
type Backend struct {
	conf config.Config
	log  *logger.Logger
}

// Submit submits a task. For the Local backend this results in the task
// running immediately.
func (b *Backend) Submit(task *tes.Task) error {
	w, err := worker.NewDefaultWorker(b.conf.Worker, task.Id, b.log)
	if err != nil {
		return err
	}
	go w.Run(context.Background())
	return nil
}
