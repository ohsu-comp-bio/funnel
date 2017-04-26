package gce

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	gce_mocks "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
	"github.com/ohsu-comp-bio/funnel/tests"
)

func init() {
	logger.ForceColors()
}

type harness struct {
	conf       config.Config
	srv        *tests.Funnel
	gceClient  Client
	mockClient *gce_mocks.Client
}

func (h *harness) Schedule() {
	h.srv.Scheduler.Schedule(context.Background())
}

func (h *harness) Scale() {
	h.srv.Scheduler.Scale(context.Background())
}

func setup() *harness {
	conf := tests.NewConfig()
	conf.Scheduler = "gce-mock"
	conf.Backends.GCE.Project = "test-proj"
	conf.Backends.GCE.Zone = "test-zone"

	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.Client)

	// Mock the server/database so we can easily control available workers
	srv := tests.NewFunnel(conf)

	h := &harness{conf, srv, gce, gce}

	// Add mock backend
	h.srv.Scheduler.AddBackend(&scheduler.BackendPlugin{
		Name: "gce-mock",
		Create: func(conf config.Config) (scheduler.Backend, error) {
			log.Debug("Creating mock scheduler backend")
			b := &Backend{conf, h.srv.Client(), h.gceClient}
			return scheduler.Backend(b), nil
		},
	})
	h.srv.Start()

	return h
}

func testWorker(id string, s pbf.WorkerState) *pbf.Worker {
	return &pbf.Worker{
		Id: id,
		Resources: &pbf.Resources{
			Cpus: 10.0,
			Ram:  100.0,
			Disk: 1000.0,
		},
		Available: &pbf.Resources{
			Cpus: 10.0,
			Ram:  100.0,
			Disk: 1000.0,
		},
		Zone:  "ok-zone",
		State: s,
		Metadata: map[string]string{
			"gce": "yes",
		},
	}
}
