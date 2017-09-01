package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

var log = logger.Sub("scheduler")

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
	log.Debug("Submitting to funnel scheduler", "taskID", task.Id)
	err := b.db.QueueTask(task)
	if err != nil {
		log.Error("Failed to submit task to the scheduler queue", "error", err, "taskID", task.Id)
		return err
	}
	return nil
}
