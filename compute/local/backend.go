package local

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.Config, log *logger.Logger, fac scheduler.Worker) *Backend {
	return &Backend{conf, log, fac}
}

// Backend represents the local backend.
type Backend struct {
	conf      config.Config
	log       *logger.Logger
	newWorker scheduler.Worker
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ev.GetTask())
	}
	return nil
}

// Submit submits a task. For the Local backend this results in the task
// running immediately.
func (b *Backend) Submit(task *tes.Task) error {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		b.newWorker(ctx, b.conf.Worker, task.Id, b.log)
	}()
	return nil
}
