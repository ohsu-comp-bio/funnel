package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NoopWorker is useful during testing for creating a worker with a Worker
// that doesn't do anything.
type NoopWorker struct {
	OnRun func(context.Context, *tes.Task)
}

// Run doesn't do anything, it's an empty function.
func (n *NoopWorker) Run(ctx context.Context, t *tes.Task) {
	if n.OnRun != nil {
		n.OnRun(ctx, t)
	}
}
