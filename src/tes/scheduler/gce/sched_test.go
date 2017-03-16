package gce

import (
	"github.com/stretchr/testify/mock"
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	gce_mocks "tes/scheduler/gce/mocks"
	server_mocks "tes/server/mocks"
	pbr "tes/server/proto"
	"tes/worker"
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

func testWorker(id string, s pbr.WorkerState) *pbr.Worker {
	return &pbr.Worker{
		Id: id,
		Resources: &pbr.Resources{
			Cpus: 10.0,
			Ram:  100.0,
			Disk: 1000.0,
		},
		Available: &pbr.Resources{
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

// TestSchedToExisting tests the case where an existing worker has capacity
// available for the task. In this case, there are no instance templates,
// so the scheduler will not create any new workers.
func TestSchedToExisting(t *testing.T) {

	// Mock config
	conf := basicConf()
	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.Client)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	// Represents a worker that is alive but at full capacity
	existing := testWorker("existing", pbr.WorkerState_Alive)
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

	// Add an instance template to the config. The scheduler uses these templates
	// to start new worker instances.
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")

	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.Client)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	// Represents a worker that is alive but at full capacity
	existing := testWorker("existing", pbr.WorkerState_Alive)
	existing.Resources.Cpus = 0.0
	srv.AddWorker(existing)

	srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

	// Mock an instance template response with 1 cpu/ram/disk
	gce.On("Template", "test-proj", "test-zone", "test-tpl").Return(&pbr.Resources{
		Cpus: 10.0,
		Ram:  100.0,
		Disk: 1000.0,
	}, nil)

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 2 {
		log.Debug("Workers", workers)
		t.Error("Expected new worker to be added to database")
	}

	expected := workers[1]
	log.Debug("Workers", workers)

	// Expected worker config
	wconf := conf.Worker
	// Expect ServerAddress to match the server's config
	wconf.ServerAddress = conf.HostName + ":" + conf.RPCPort
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
	gce := new(gce_mocks.Client)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	// Represents a worker that is alive but at full capacity
	existing := testWorker("existing", pbr.WorkerState_Alive)
	existing.Resources.Cpus = 10.0
	srv.AddWorker(existing)

	srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

	// Mock an instance template response with 1 cpu/ram/disk
	gce.On("Template", "test-proj", "test-zone", "test-tpl").Return(&pbr.Resources{
		Cpus: 10.0,
		Ram:  100.0,
		Disk: 1000.0,
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

// Test submit multiple jobs at once when no workers exist. Multiple workers
// should be started.
func TestSchedStartMultipleWorker(t *testing.T) {

	// Mock config
	conf := basicConf()
	// Add an instance template to the config. The scheduler uses these templates
	// to start new worker instances.
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")

	// Mock the GCE API so actual API calls aren't needed
	gce := new(gce_mocks.Client)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	srv.RunHelloWorld()
	srv.RunHelloWorld()
	srv.RunHelloWorld()
	srv.RunHelloWorld()

	// The GCE scheduler under test
	s := &gceScheduler{conf, srv.Client, gce}

	// Mock an instance template response with 1 cpu/ram
	gce.On("Template", "test-proj", "test-zone", "test-tpl").Return(&pbr.Resources{
		Cpus: 1.0,
		Ram:  1.0,
		Disk: 1000.0,
	}, nil)

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 4 {
		log.Debug("WORKERS", workers)
		t.Error("Expected multiple workers")
	}
}

// Test that assigning a job to a worker correctly updates the available resources.
func TestUpdateAvailableResources(t *testing.T) {

	conf := basicConf()
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	existing := testWorker("existing", pbr.WorkerState_Alive)
	srv.AddWorker(existing)
	ta := srv.HelloWorldTask()
	srv.RunTask(ta)
	s := &gceScheduler{conf, srv.Client, new(gce_mocks.Client)}

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 1 || workers[0].Id != "existing" {
		log.Debug("WORKERS", workers)
		t.Error("Expected a single, existing worker")
	}

	expect := existing.Resources.Cpus - ta.Resources.MinimumCpuCores

	if workers[0].Available.Cpus != expect {
		t.Error("Unexpected cpu count")
	}
}

// Try to reproduce a bug where available CPUs seems to overflow
func TestUpdateBugAvailableResources(t *testing.T) {

	conf := basicConf()
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	existingA := testWorker("existing-A", pbr.WorkerState_Alive)
	existingA.Resources.Cpus = 8.0
	srv.AddWorker(existingA)

	existingB := testWorker("existing-B", pbr.WorkerState_Alive)
	existingB.Resources.Cpus = 8.0
	srv.AddWorker(existingB)

	ta := srv.HelloWorldTask()
	tb := srv.HelloWorldTask()
	tc := srv.HelloWorldTask()
	ta.Resources.MinimumCpuCores = 4
	tb.Resources.MinimumCpuCores = 4
	tc.Resources.MinimumCpuCores = 4
	srv.RunTask(ta)
	srv.RunTask(tb)
	srv.RunTask(tc)

	s := &gceScheduler{conf, srv.Client, new(gce_mocks.Client)}

	scheduler.ScheduleChunk(srv.DB, s, conf)
	workers := srv.GetWorkers()

	log.Debug("WORKERS", workers)

	if len(workers) != 2 {
		t.Error("Expected a single, existing worker")
	}

	tot := workers[0].Available.Cpus + workers[1].Available.Cpus

	if tot != 4 {
		t.Error("Expected total available cpu count to be 4")
	}
}

// Test a bug where worker resources were not being correctly reported/updated,
// causing jobs to be scheduled incorrectly.
func TestSchedMultipleJobsResourceUpdateBug(t *testing.T) {

	conf := basicConf()
	gce := new(gce_mocks.Client)
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")
	srv := server_mocks.MockServerFromConfig(conf)
	defer srv.Close()
	s := &gceScheduler{srv.Conf, srv.Client, gce}

	var w *worker.Worker

	// Mock an instance template response with 1 cpu/ram/disk
	gce.On("Template", "test-proj", "test-zone", "test-tpl").Return(&pbr.Resources{
		Cpus: 10.0,
		Ram:  100.0,
		Disk: 1000.0,
	}, nil)

	// This test stems from a bug found during testing GCE worker init.
	// Mock out a started worker to match the scenario the bug was found.
	//
	// The root problem was that the scheduler could schedule one job but not two,
	// because the Disk resources would first be reported by the GCE instance template,
	// but once the worker sent an update, the resource information was incorrectly
	// reported and merged. This test tries to replicate that scenario closely.
	gce.
		On("StartWorker", "test-proj", "test-zone", "test-tpl", mock.Anything).
		Run(func(args mock.Arguments) {
			wconf := args[3].(config.Worker)
			w = newMockWorker(wconf)
		}).
		Return(nil)

		// Run the first task. This will be scheduled.
	ida := srv.RunHelloWorld()
	scheduler.ScheduleChunk(srv.DB, s, srv.Conf)
	// Starts the worker.
	scheduler.Scale(srv.DB, s)

	// Sync the worker state.
	w.Sync()
	// Mark the first task as complete without error
	w.Ctrls[ida].SetResult(nil)
	w.Sync()

	log.Debug("RESOURCES", srv.GetWorkers())

	// Run the second task. This wasn't being scheduled, which is a bug.
	idb := srv.RunHelloWorld()
	scheduler.ScheduleChunk(srv.DB, s, srv.Conf)

	log.Debug("RESOURCES", srv.GetWorkers())

	w.Sync()
	if _, ok := w.Ctrls[idb]; !ok {
		t.Error("The second job didn't get to the worker as expected")
	}

	if len(srv.GetWorkers()) != 1 {
		t.Error("Expected only one worker")
	}
}

func newMockWorker(conf config.Worker) *worker.Worker {
	w, err := worker.NewWorker(conf)
	if err != nil {
		panic(err)
	}
	w.JobRunner = worker.NoopJobRunner
	w.Sync()
	return w
}
