package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Worker is a type which runs a task.
type Worker interface {
	Run(context.Context)
}

// TaskReader is a type which reads and writes task information
// during task execution.
type TaskReader interface {
	Task() (*tes.Task, error)
	State() tes.State
}
