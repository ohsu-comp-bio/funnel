package gce

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	gcemock "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
	schedmock "github.com/ohsu-comp-bio/funnel/scheduler/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestPreferExisting(t *testing.T) {
	sched := new(schedmock.Client)
	gce := new(gcemock.Client)

	// Set up data for an existing (ALIVE state) node,
	// and a template (UNINITIALIZED state) node.
	w := pbs.Node{
		Resources: &pbs.Resources{
			Cpus:   10,
			RamGb:  100.0,
			DiskGb: 100.0,
		},
		Available: &pbs.Resources{
			Cpus:   10,
			RamGb:  100.0,
			DiskGb: 100.0,
		},
		Metadata: map[string]string{"gce": "yes"},
	}
	existing := w
	existing.Id = "existing"
	existing.State = pbs.NodeState_ALIVE
	template := w
	template.Id = "template"

	// Return existing and template from mock API clients.
	sched.On("ListNodes", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbs.ListNodesResponse{
			Nodes: []*pbs.Node{&existing},
		}, nil)

	gce.On("Templates").Return([]pbs.Node{template})

	b := Backend{
		client: sched,
		gce:    gce,
	}

	// Call schedule many times, to ensure the result is consistent.
	for i := 0; i < 100; i++ {
		o := b.Schedule(&tes.Task{})
		if o == nil || o.Node.Id != "existing" {
			logger.Debug("", "offer", o, "i", i)
			t.Fatalf("expected schedule to return existing node")
		}
	}
}
