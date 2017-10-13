package worker

import (
	"context"
)

// NoopWorker is useful during testing for creating a worker with a Worker
// that doesn't do anything.
type NoopWorker struct{}

// Run doesn't do anything, it's an empty function.
func (NoopWorker) Run(context.Context) {}

func (NoopWorker) Close() error {
	return nil
}
