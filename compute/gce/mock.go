package gce

import (
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
)

// MockBackend is a GCE backend that doesn't communicate with
// Google Cloud APIs, which is useful for testing.
type MockBackend struct {
	*Backend
	Wrapper
}

// NewMockBackend returns a GCE scheduler backend that doesn't
// communicate with Google Cloud APIs,
// Useful for testing.
func NewMockBackend(conf config.Config, w Wrapper) (*MockBackend, error) {
	// Set up a GCE scheduler backend that has a mock client
	// so that it doesn't actually communicate with GCE.

	gceClient := &gceClient{
		wrapper: w,
		project: conf.Backends.GCE.Project,
		zone:    conf.Backends.GCE.Zone,
	}

	wpClient, err := scheduler.NewClient(conf.Scheduler.Node.RPC)
	if err != nil {
		return nil, err
	}

	return &MockBackend{
		Backend: &Backend{
			conf:   conf,
			client: wpClient,
			gce:    gceClient,
		},
		Wrapper: w,
	}, nil
}
