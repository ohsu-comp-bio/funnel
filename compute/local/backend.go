// Package local contains code for accessing compute resources via the local computer, for Funnel development and debugging.
package local

import (
	"context"
	"syscall"
	"time"

	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, conf config.Config, log *logger.Logger) (*Backend, error) {
	return &Backend{conf, log}, nil
}

// Backend represents the local backend.
type Backend struct {
	conf config.Config
	log  *logger.Logger
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
	ctx, cancel := context.WithCancel(context.Background())
	ctx = util.SignalContext(ctx, time.Millisecond, syscall.SIGINT, syscall.SIGTERM)

	w, err := workerCmd.NewWorker(ctx, b.conf, b.log, &workerCmd.WorkerOpts{
		TaskID: task.Id,
	})
	if err != nil {
		return err
	}

	go func() {
		defer cancel()
		w.Run(ctx)
	}()
	return nil
}
