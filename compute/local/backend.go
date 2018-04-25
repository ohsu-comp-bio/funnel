// Package local contains code for accessing compute resources via the local computer, for Funnel development and debugging.
package local

import (
	"context"
	"syscall"
	"time"

	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
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

// Submit submits a task. For the Local backend this results in the task
// running immediately.
func (b *Backend) CreateTask(ctx context.Context, task *tes.Task) error {
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		ctx = util.SignalContext(ctx, time.Millisecond, syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		b.worker.Run(ctx, task.Id)
	}()
	return nil
}

func (b *Backend) CancelTask(ctx context.Context, id string) error {
	return nil
}
