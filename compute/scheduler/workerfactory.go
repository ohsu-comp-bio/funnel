package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// Worker is a function which creates a new worker instance.
type Worker func(ctx context.Context, c config.Config, taskID string, log *logger.Logger) error

// NoopWorker does nothing.
func NoopWorker(ctx context.Context, c config.Config, taskID string, log *logger.Logger) error {
	return nil
}
