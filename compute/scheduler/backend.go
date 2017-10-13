package scheduler

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewComputeBackend returns a new scheduler ComputeBackend instance.
func NewComputeBackend(db Database) *ComputeBackend {
	return &ComputeBackend{db}
}

// ComputeBackend represents the funnel scheduler backend.
type ComputeBackend struct {
	db Database
}

// Submit submits a task via gRPC call to the funnel scheduler backend
func (b *ComputeBackend) Submit(task *tes.Task) error {
	err := b.db.QueueTask(task)
	if err != nil {
		return fmt.Errorf("Failed to submit task %s to the scheduler queue: %s", task.Id, err)
	}
	return nil
}
