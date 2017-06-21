package noop

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// Name of the the scheduler backend
const Name = "noop"

// NewBackend returns a new noop scheduler backend.
// A noop backend uses a worker that doesn't have any side effects,
// (e.g. file access, docker calls, etc) which is useful for testing.
func NewBackend(conf config.Config) (*Backend, error) {
	w, err := worker.NewNoopWorker(conf.Worker)
	if err != nil {
		return nil, err
	}
	return &Backend{w, conf}, nil
}

// Backend is a scheduler backend which doesn't have any side effects,
// (e.g. file access, docker calls, etc) which is useful for testing.
type Backend struct {
	Worker *worker.Worker
	conf   config.Config
}

// Schedule schedules a task to the noop worker. There is only
// one worker and tasks are always scheduled to that worker without
// any logic or filtering (just dead simple).
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	w := &pbf.Worker{
		Id:    s.conf.Worker.ID,
		State: pbf.WorkerState_ALIVE,
	}
	return scheduler.NewOffer(w, j, scheduler.Scores{})
}
