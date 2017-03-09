package gce

import (
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	gce_mocks "tes/scheduler/gce/mocks"
	server_mocks "tes/server/mocks"
	pbr "tes/server/proto"
	"testing"
)

func init() {
	logger.ForceColors()
}

func basicConf() config.Config {
	conf := config.DefaultConfig()
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")
	conf.Schedulers.GCE.Project = "test-proj"
	conf.Schedulers.GCE.Zone = "test-zone"
	return conf
}

func worker(id string, s pbr.WorkerState) *pbr.Worker {
	return &pbr.Worker{
		Id: id,
		Resources: &pbr.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
		},
		Available: &pbr.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
		},
		Zone:  "ok-zone",
		State: s,
		Metadata: map[string]string{
			"gce": "yes",
		},
	}
}

/*
func TestSchedBasic(t *testing.T) {
	existing := worker("existing", pbr.WorkerState_Alive)
	template := worker("template", pbr.WorkerState_Uninitialized)

	conf := config.DefaultConfig()
	gcemock, tesmock := newMocks()
	gcemock.templates = append(gcemock.templates, template)
	tesmock.workers = append(tesmock.workers, existing)

	s := gceScheduler{conf, tesmock, gcemock}
	j := &pbe.Job{}
	r := s.Schedule(j)
	if r == nil {
		t.Error("Job was not scheduled")
	}
	if r.Worker.Id != "existing" {
		log.Debug("Worker", r.Worker)
		t.Error("Job was scheduled to wrong worker")
	}
}
*/

// TestSchedStartWorker tests the case where the scheduler wants to start a new
// GCE worker instance from a instance template defined in the configuration.
// The scheduler calls the GCE API to get the template details and assigns
// a job to that unintialized worker. The scaler then calls the GCE API to
// start the worker.
func TestSchedStartWorker(t *testing.T) {

  // Mock config
	conf := basicConf()
	// Set a different server address to test that it gets passed on to the worker
	conf.ServerAddress = "other:9090"

  // Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.GCEClient)
  // Mock the server/database so we can easily control available workers
  srv := server_mocks.NewMockServer()

	// Represents a worker that is alive but at full capacity
	existing := worker("existing", pbr.WorkerState_Alive)
	existing.Resources.Cpus = 0.0
  srv.AddWorker(existing)

  srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

  // Mock an instance template response with 1 cpu/ram/disk
	gce.On("Template", "test-proj", "test-tpl").Return(&pbr.Resources{
		Cpus: 1.0,
		Ram:  1.0,
		Disk: 1.0,
	}, nil)

	scheduler.ScheduleChunk(srv.DB, s, conf)
  workers := srv.GetWorkers()

  if len(workers) != 2 {
    t.Error("Expected new worker to be added to database")
  }

  expected := workers[1]
  log.Debug("Workers", workers)

	// Expected worker config
	wconf := conf.Worker
	// Expect ServerAddress to match the server's config
	wconf.ServerAddress = conf.ServerAddress
	wconf.ID = expected.Id
	gce.On("StartWorker", "test-proj", "test-zone", "test-tpl", wconf).Return(nil)

	scheduler.Scale(srv.DB, s)
	gce.AssertExpectations(t)
}

/*
func TestGetWorkers(t *testing.T) {
	existing := worker("existing", pbr.WorkerState_Alive)
	existing.Available.Cpus = 0.0
	template := worker("template", pbr.WorkerState_Uninitialized)

	conf := config.DefaultConfig()
	gcemock, tesmock := newMocks()
	gcemock.templates = append(gcemock.templates, template)
	tesmock.workers = append(tesmock.workers, existing)

	s := gceScheduler{conf, tesmock, gcemock}
	r := s.getWorkers()

	if len(r) != 2 {
		t.Error("Expected 2 workers")
	}
	if r[0] != existing {
		log.Debug("Worker", r[0])
		t.Error("Unexpected worker")
	}
  if r[1].Id != "template" {
		log.Debug("Worker", r[1])
		t.Error("Unexpected second worker")
	}
}
*/
