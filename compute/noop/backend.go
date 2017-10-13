package noop

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new noop Backend instance.
func NewBackend(conf config.Config) *Backend {
	return &Backend{conf}
}

// Backend is a scheduler backend that doesn't do anything
// which is useful for testing.
type Backend struct {
	conf config.Config
}

// Submit submits a task. For the noop backend this does nothing.
func (b *Backend) Submit(task *tes.Task) error {
	return nil
}
