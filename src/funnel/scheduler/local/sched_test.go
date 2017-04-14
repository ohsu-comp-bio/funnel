package local

import (
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	sched "funnel/scheduler"
	sched_mocks "funnel/scheduler/mocks"
	. "github.com/stretchr/testify/mock"
	"testing"
)

func init() {
	logger.ForceColors()
}

func simpleWorker() *pbf.Worker {
	return &pbf.Worker{
		// This ID MUST match the ID set in setup()
		// because the local scheduler is built to have only a single worker
		Id: "test-worker-id",
		Resources: &pbf.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
		},
		Available: &pbf.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
		},
		State: pbf.WorkerState_Alive,
		Zone:  "ok-zone",
	}
}

func setup(workers []*pbf.Worker) (*sched_mocks.Client, *Backend) {
	conf := config.Config{}
	mc := new(sched_mocks.Client)

	// Mock in test workers
	mc.On("GetWorkers", Anything, Anything, Anything).Return(&pbf.GetWorkersResponse{
		Workers: workers,
	}, nil)

	s := &Backend{
		conf,
		mc,
		"test-worker-id",
	}
	return mc, s
}

func TestNoWorkers(t *testing.T) {
	_, s := setup([]*pbf.Worker{})
	j := &tes.Task{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Task scheduled on empty workers")
	}
}

func TestSingleWorker(t *testing.T) {
	_, s := setup([]*pbf.Worker{
		simpleWorker(),
	})

	j := &tes.Task{}
	o := s.Schedule(j)
	if o == nil {
		t.Error("Failed to schedule task on single worker")
		return
	}
	if o.Worker.Id != "test-worker-id" {
		t.Error("Scheduled task on unexpected worker")
	}
}

// Test that the scheduler ignores workers it doesn't own.
func TestIgnoreOtherWorkers(t *testing.T) {
	other := simpleWorker()
	other.Id = "other-worker"

	_, s := setup([]*pbf.Worker{other})

	j := &tes.Task{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to other worker")
	}
}

// Test that scheduler ignores workers without the "Alive" state
func TestIgnoreNonAliveWorkers(t *testing.T) {
	j := &tes.Task{}

	for name, val := range pbf.WorkerState_value {
		w := simpleWorker()
		w.State = pbf.WorkerState(val)
		_, s := setup([]*pbf.Worker{w})
		o := s.Schedule(j)

		if name == "Alive" {
			// Testing Alive just so I know this test is worker as expected
			if o == nil {
				t.Error("Didn't schedule task to alive worker")
			}
		} else {
			if o != nil {
				t.Errorf("Scheduled task to non-alive worker: %s", name)
				return
			}
		}
	}
}

// Test whether the scheduler correctly filters workers based on
// cpu, ram, disk, etc.
func TestMatch(t *testing.T) {
	_, s := setup([]*pbf.Worker{
		simpleWorker(),
	})

	var o *sched.Offer
	var j *tes.Task

	// Helper which sets up Task.Resources struct to non-nil
	blankTask := func() *tes.Task {
		return &tes.Task{Task: &tes.Task{Resources: &tes.Resources{}}}
	}

	// test CPUs too big
	j = blankTask()
	j.Task.Resources.MinimumCpuCores = 2
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to worker without enough CPU resources")
	}

	// test RAM too big
	j = blankTask()
	j.Task.Resources.MinimumRamGb = 2.0
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to worker without enough RAM resources")
	}

	// test disk too big
	j = blankTask()
	j.Task.Resources.Volumes = []*tes.Volume{
		{SizeGb: 2.0},
	}

	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to worker without enough Disk resources")
	}

	// test two volumes, basically check that they are
	// added together to get total size
	j = blankTask()
	j.Task.Resources.Volumes = []*tes.Volume{
		{SizeGb: 1.0},
		{SizeGb: 0.1},
	}

	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to worker without enough Disk resources, 2 volumes")
	}

	// test zones don't match
	j = blankTask()
	j.Task.Resources.Zones = []string{"test-zone"}
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to worker out of zone")
	}

	// Now test a task that fits
	j = blankTask()
	j.Task.Resources.MinimumCpuCores = 1
	j.Task.Resources.MinimumRamGb = 1.0
	j.Task.Resources.Volumes = []*tes.Volume{
		{SizeGb: 0.5},
		{SizeGb: 0.5},
	}
	j.Task.Resources.Zones = []string{"ok-zone", "not-ok-zone"}

	o = s.Schedule(j)
	if o == nil {
		t.Error("Didn't schedule task when resources fit")
	}
}
