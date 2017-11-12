package worker

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// TaskReader is a type which reads task information
// during task execution.
type TaskReader interface {
	Task() (*tes.Task, error)
	State() (tes.State, error)
}
