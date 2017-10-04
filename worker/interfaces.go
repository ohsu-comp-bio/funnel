package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Worker is a type which runs a task.
type Worker interface {
	Run(context.Context, *tes.Task)
}
