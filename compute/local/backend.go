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
func NewBackend(conf config.Config, fac scheduler.Worker, log *logger.Logger) *Backend {
	return &Backend{conf, fac, log}
}

// Backend represents the local backend.
type Backend struct {
	conf      config.Config
	newWorker scheduler.Worker
	log       *logger.Logger
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
		err := b.newWorker(ctx, b.conf, task.Id, b.log)
		if err != nil {
			b.log.Error("failed to run task", err)
		}
	}()
	return nil
}
