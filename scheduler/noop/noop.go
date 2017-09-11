package noop

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
)

var log = logger.Sub("noop")

// NewBackend returns a new noop scheduler backend.
// A noop backend uses a node that doesn't have any side effects,
// (e.g. file access, docker calls, etc) which is useful for testing.
func NewBackend(conf config.Config) (*Backend, error) {
	n, err := scheduler.NewNoopNode(conf)
	if err != nil {
		log.Error("Failed to create noop node", err)
		return nil, err
	}
	return &Backend{n, conf}, nil
}

// Backend is a scheduler backend which doesn't have any side effects,
// (e.g. file access, docker calls, etc) which is useful for testing.
type Backend struct {
	Node *scheduler.Node
	conf config.Config
}

// Schedule schedules a task to the noop node. There is only
// one node and tasks are always scheduled to that node without
// any logic or filtering (just dead simple).
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running noop scheduler")

	n := &pbs.Node{
		Id:    s.conf.Scheduler.Node.ID,
		State: pbs.NodeState_ALIVE,
	}
	return scheduler.NewOffer(n, j, scheduler.Scores{})
}
