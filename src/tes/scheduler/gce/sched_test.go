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

// TestSchedToExisting tests the case where an existing worker has capacity
// available for the task. In this case, there are no instance templates,
// so the scheduler will not create any new workers.
func TestSchedToExisting(t *testing.T) {

	// Mock config
	conf := basicConf()
	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.GCEClient)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	// Represents a worker that is alive but at full capacity
	existing := worker("existing", pbr.WorkerState_Alive)
	existing.Resources.Cpus = 0.0
	srv.AddWorker(existing)
	srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 1 {
		t.Error("Expected a single worker")
	}

	log.Debug("Workers", workers)
	w := workers[0]

	if w.Id != "existing" {
		t.Error("Job scheduled to unexpected worker")
	}

	// Not really needed for this test, but safer anyway
	scheduler.Scale(srv.DB, s)
	// Basically asserts that no GCE APIs were called. None are needed in this case.
	gce.AssertExpectations(t)
}

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
	// Add an instance template to the config. The scheduler uses these templates
	// to start new worker instances.
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")

	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.GCEClient)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

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

// TestPreferExistingWorker tests the case where there is an existing worker
// AND instance templates available. The existing worker has capacity for the task,
// and the task should be scheduled to the existing worker.
func TestPreferExistingWorker(t *testing.T) {

	// Mock config
	conf := basicConf()
	// Add an instance template to the config. The scheduler uses these templates
	// to start new worker instances.
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")

	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.GCEClient)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	// Represents a worker that is alive but at full capacity
	existing := worker("existing", pbr.WorkerState_Alive)
	existing.Resources.Cpus = 10.0
	srv.AddWorker(existing)

	srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

	// Mock an instance template response with 1 cpu/ram/disk
	gce.On("Template", "test-proj", "test-tpl").Return(&pbr.Resources{
		Cpus: 10.0,
		Ram:  1.0,
		Disk: 1.0,
	}, nil)

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 1 {
		t.Error("Expected no new workers to be created")
	}

	expected := workers[0]
	log.Debug("Workers", workers)

	if expected.Id != "existing" {
		t.Error("Job was scheduled to the wrong worker")
	}

	// Nothing should be scaled in this test, but safer to call Scale anyway
	scheduler.Scale(srv.DB, s)
	gce.AssertExpectations(t)
}
