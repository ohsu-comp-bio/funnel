package noop

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
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

// WriteEvent is a noop and returns nil.
func (b *Backend) WriteEvent(context.Context, *events.Event) error {
	return nil
}
