package gce

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	gcemock "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
	schedmock "github.com/ohsu-comp-bio/funnel/scheduler/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestPreferExisting(t *testing.T) {
	sched := new(schedmock.Client)
	gce := new(gcemock.Client)

	// Set up data for an existing (ALIVE state) worker,
	// and a template (UNINITIALIZED state) worker.
	w := pbf.Worker{
		Resources: &pbf.Resources{
			Cpus:   10,
			RamGb:  100.0,
			DiskGb: 100.0,
		},
		Available: &pbf.Resources{
			Cpus:   10,
			RamGb:  100.0,
			DiskGb: 100.0,
		},
		Metadata: map[string]string{"gce": "yes"},
	}
	existing := w
	existing.Id = "existing"
	existing.State = pbf.WorkerState_ALIVE
	template := w
	template.Id = "template"

	// Return existing and template from mock API clients.
	sched.On("ListWorkers", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.ListWorkersResponse{
			Workers: []*pbf.Worker{&existing},
		}, nil)

	gce.On("Templates").Return([]pbf.Worker{template})

	b := Backend{
		client: sched,
		gce:    gce,
	}

	// Call schedule many times, to ensure the result is consistent.
	for i := 0; i < 100; i++ {
		o := b.Schedule(&tes.Task{})
		if o == nil || o.Worker.Id != "existing" {
			logger.Debug("", "offer", o, "i", i)
			t.Fatalf("expected schedule to return existing worker")
		}
	}
}
