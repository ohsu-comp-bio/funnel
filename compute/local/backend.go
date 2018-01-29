package local

import (
	"context"
	"syscall"
	"time"

	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, conf config.Config, log *logger.Logger) (*Backend, error) {
	w, err := workerCmd.NewWorker(ctx, conf, log)
	if err != nil {
		return nil, err
	}
	return &Backend{conf, w, log}, nil
}

// Backend represents the local backend.
type Backend struct {
	conf   config.Config
	worker *worker.DefaultWorker
	log    *logger.Logger
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
		ctx = util.SignalContext(ctx, time.Millisecond, syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		b.worker.Run(ctx, task.Id)
	}()
	return nil
}
