package worker

import (
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// TaskReader is a type which reads task information
// during task execution.
type TaskReader interface {
	Task(ctx context.Context) (*tes.Task, error)
	State(ctx context.Context) (tes.State, error)
}
