package gce

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/node"
	gcemock "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
)

// MockBackend is a GCE backend that doesn't communicate with
// Google Cloud APIs, which is useful for testing.
type MockBackend struct {
	*Backend
	Wrapper *gcemock.Wrapper
}

// NewMockBackend returns a GCE scheduler backend that doesn't
// communicate with Google Cloud APIs,
// Useful for testing.
func NewMockBackend(conf config.Config) (*MockBackend, error) {
	// Set up a GCE scheduler backend that has a mock client
	// so that it doesn't actually communicate with GCE.

	gceWrapper := new(gcemock.Wrapper)
	gceClient := &gceClient{
		wrapper: gceWrapper,
		project: conf.Backends.GCE.Project,
		zone:    conf.Backends.GCE.Zone,
	}

	wpClient, err := node.NewClient(conf.Scheduler.Node)
	if err != nil {
		return nil, err
	}

	return &MockBackend{
		Backend: &Backend{
			conf:   conf,
			client: wpClient,
			gce:    gceClient,
		},
		Wrapper: gceWrapper,
	}, nil
}
