package gce

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/worker"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestSchedToExisting tests the case where an existing worker has capacity
// available for the task. In this case, there are no instance templates,
// so the scheduler will not create any new workers.
func TestSchedToExisting(t *testing.T) {
	h := setup()
	defer h.srv.Stop()

	existing := testWorker("existing", pbf.WorkerState_Alive)
	h.mockClient.SetupEmptyMockTemplates()
	h.srv.AddWorker(existing)
	h.srv.RunHelloWorld()

	h.Schedule()
	workers := h.srv.GetWorkers()

	if len(workers) != 1 {
		t.Error("Expected a single worker")
	}

	log.Debug("Workers", workers)
	w := workers[0]

	if w.Id != "existing" {
		t.Error("Task scheduled to unexpected worker")
	}

	// Not really needed for this test, but safer anyway
	h.Scale()
	// Basically asserts that no GCE APIs were called. None are needed in this case.
	h.mockClient.AssertExpectations(t)
}

// TestSchedStartWorker tests the case where the scheduler wants to start a new
// GCE worker instance from a instance template defined in the configuration.
// The scheduler calls the GCE API to get the template details and assigns
// a task to that unintialized worker. The scaler then calls the GCE API to
// start the worker.
func TestSchedStartWorker(t *testing.T) {
	h := setup()
	defer h.srv.Stop()

	h.mockClient.SetupDefaultMockTemplates()

	// Represents a worker that is alive but at full capacity
	existing := testWorker("existing", pbf.WorkerState_Alive)
	existing.Resources.Cpus = 0.0
	h.srv.AddWorker(existing)

	h.srv.RunHelloWorld()
	h.Schedule()
	workers := h.srv.GetWorkers()

	if len(workers) != 2 {
		log.Debug("Workers", workers)
		t.Error("Expected new worker to be added to database")
		return
	}

	expected := workers[1]
	log.Debug("Workers", workers)

	// Expect ServerAddress to match the server's config
	h.mockClient.On("StartWorker", "test-tpl", h.conf.RPCAddress(), expected.Id).Return(nil)

	h.Scale()
	h.mockClient.AssertExpectations(t)
}

// TestPreferExistingWorker tests the case where there is an existing worker
// AND instance templates available. The existing worker has capacity for the task,
// and the task should be scheduled to the existing worker.
func TestPreferExistingWorker(t *testing.T) {
	h := setup()
	defer h.srv.Stop()

	h.mockClient.SetupDefaultMockTemplates()

	existing := testWorker("existing", pbf.WorkerState_Alive)
	existing.Resources.Cpus = 10.0
	h.srv.AddWorker(existing)

	h.srv.RunHelloWorld()

	h.Schedule()
	workers := h.srv.GetWorkers()

	if len(workers) != 1 {
		t.Error("Expected no new workers to be created")
	}

	expected := workers[0]
	log.Debug("Workers", workers)

	if expected.Id != "existing" {
		t.Error("Task was scheduled to the wrong worker")
	}

	// Nothing should be scaled in this test, but safer to call Scale anyway
	h.Scale()
	h.mockClient.AssertExpectations(t)
}

// Test submit multiple tasks at once when no workers exist. Multiple workers
// should be started.
func TestSchedStartMultipleWorker(t *testing.T) {
	h := setup()
	defer h.srv.Stop()

	h.srv.RunHelloWorld()
	h.srv.RunHelloWorld()
	h.srv.RunHelloWorld()
	h.srv.RunHelloWorld()

	// Mock an instance template response with 1 cpu/ram
	h.mockClient.SetupMockTemplates(pbf.Resources{
		Cpus: 1.0,
		Ram:  1.0,
		Disk: 1000.0,
	})

	h.Schedule()
	workers := h.srv.GetWorkers()

	if len(workers) != 4 {
		log.Debug("WORKERS", workers)
		t.Error("Expected multiple workers")
	}
}

// Test that assigning a task to a worker correctly updates the available resources.
func TestUpdateAvailableResources(t *testing.T) {
	h := setup()
	defer h.srv.Stop()
	h.mockClient.SetupEmptyMockTemplates()

	existing := testWorker("existing", pbf.WorkerState_Alive)
	h.srv.AddWorker(existing)

	ta := h.srv.HelloWorldTask()
	h.srv.CreateTask(ta)

	h.Schedule()
	workers := h.srv.GetWorkers()

	if len(workers) != 1 || workers[0].Id != "existing" {
		log.Debug("WORKERS", workers)
		t.Error("Expected a single, existing worker")
	}

	expect := existing.Resources.Cpus - ta.Resources.CpuCores

	if workers[0].Available.Cpus != expect {
		t.Error("Unexpected cpu count")
	}
}

// Try to reproduce a bug where available CPUs seems to overflow
func TestUpdateBugAvailableResources(t *testing.T) {
	h := setup()
	defer h.srv.Stop()
	h.mockClient.SetupEmptyMockTemplates()

	existingA := testWorker("existing-A", pbf.WorkerState_Alive)
	existingA.Resources.Cpus = 8.0
	h.srv.AddWorker(existingA)

	existingB := testWorker("existing-B", pbf.WorkerState_Alive)
	existingB.Resources.Cpus = 8.0
	h.srv.AddWorker(existingB)

	ta := h.srv.HelloWorldTask()
	tb := h.srv.HelloWorldTask()
	tc := h.srv.HelloWorldTask()
	ta.Resources.CpuCores = 4
	tb.Resources.CpuCores = 4
	tc.Resources.CpuCores = 4
	h.srv.CreateTask(ta)
	h.srv.CreateTask(tb)
	h.srv.CreateTask(tc)

	h.Schedule()
	workers := h.srv.GetWorkers()

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
// causing tasks to be scheduled incorrectly.
func TestSchedMultipleTasksResourceUpdateBug(t *testing.T) {
	h := setup()
	defer h.srv.Stop()
	h.mockClient.SetupDefaultMockTemplates()

	var w *worker.Worker

	// This test stems from a bug found during testing GCE worker init.
	// Mock out a started worker to match the scenario the bug was found.
	//
	// The root problem was that the scheduler could schedule one task but not two,
	// because the Disk resources would first be reported by the GCE instance template,
	// but once the worker sent an update, the resource information was incorrectly
	// reported and merged. This test tries to replicate that scenario closely.
	h.mockClient.
		On("StartWorker", "test-tpl", h.conf.RPCAddress(), mock.Anything).
		Run(func(args mock.Arguments) {
			addr := args[1].(string)
			id := args[2].(string)
			wconf := h.conf.Worker
			wconf.ServerAddress = addr
			wconf.ID = id
			w = newMockWorker(wconf)
		}).
		Return(nil)

		// Run the first task. This will be scheduled.
	ida := h.srv.RunHelloWorld()
	h.Schedule()
	// Starts the worker.
	h.Scale()

	// Sync the worker state.
	w.Sync()
	// Mark the first task as complete without error
	w.Ctrls[ida].SetResult(nil)
	w.Sync()

	log.Debug("RESOURCES", h.srv.GetWorkers())

	// Run the second task. This wasn't being scheduled, which is a bug.
	idb := h.srv.RunHelloWorld()
	h.Schedule()

	log.Debug("RESOURCES", h.srv.GetWorkers())

	w.Sync()
	if _, ok := w.Ctrls[idb]; !ok {
		t.Error("The second task didn't get to the worker as expected")
	}

	if len(h.srv.GetWorkers()) != 1 {
		t.Error("Expected only one worker")
	}
}

func newMockWorker(conf config.Worker) *worker.Worker {
	w, err := worker.NewWorker(conf)
	if err != nil {
		panic(err)
	}
	w.TaskRunner = worker.NoopTaskRunner
	w.Sync()
	return w
}
