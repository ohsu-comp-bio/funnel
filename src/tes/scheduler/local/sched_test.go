package local

import (
	. "github.com/stretchr/testify/mock"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	sched_mocks "tes/scheduler/mocks"
	pbr "tes/server/proto"
	"testing"
)

func init() {
	logger.ForceColors()
}

func simpleWorker() *pbr.Worker {
	return &pbr.Worker{
		// This ID MUST match the ID set in setup()
		// because the local scheduler is built to have only a single worker
		Id: "test-worker-id",
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
		State: pbr.WorkerState_Alive,
		Zone:  "ok-zone",
	}
}

func setup(workers []*pbr.Worker) (*sched_mocks.Client, *scheduler) {
	conf := config.Config{}
	mc := new(sched_mocks.Client)

	// Mock in test workers
	mc.On("GetWorkers", Anything, Anything, Anything).Return(&pbr.GetWorkersResponse{
		Workers: workers,
	}, nil)

	s := &scheduler{
		conf,
		mc,
		"test-worker-id",
	}
	return mc, s
}

func TestNoWorkers(t *testing.T) {
	_, s := setup([]*pbr.Worker{})
	j := &pbe.Job{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Job scheduled on empty workers")
	}
}

func TestSingleWorker(t *testing.T) {
	_, s := setup([]*pbr.Worker{
		simpleWorker(),
	})

	j := &pbe.Job{}
	o := s.Schedule(j)
	if o == nil {
		t.Error("Failed to schedule job on single worker")
		return
	}
	if o.Worker.Id != "test-worker-id" {
		t.Error("Scheduled job on unexpected worker")
	}
}

// Test that the scheduler ignores workers it doesn't own.
func TestIgnoreOtherWorkers(t *testing.T) {
	other := simpleWorker()
	other.Id = "other-worker"

	_, s := setup([]*pbr.Worker{other})

	j := &pbe.Job{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to other worker")
	}
}

// Test that scheduler ignores workers without the "Alive" state
func TestIgnoreNonAliveWorkers(t *testing.T) {
	j := &pbe.Job{}

	for name, val := range pbr.WorkerState_value {
		w := simpleWorker()
		w.State = pbr.WorkerState(val)
		_, s := setup([]*pbr.Worker{w})
		o := s.Schedule(j)

		if name == "Alive" {
			// Testing Alive just so I know this test is worker as expected
			if o == nil {
				t.Error("Didn't schedule job to alive worker")
			}
		} else {
			if o != nil {
				t.Errorf("Scheduled job to non-alive worker: %s", name)
				return
			}
		}
	}
}

// Test whether the scheduler correctly filters workers based on
// cpu, ram, disk, etc.
func TestMatch(t *testing.T) {
	_, s := setup([]*pbr.Worker{
		simpleWorker(),
	})

	var o *sched.Offer
	var j *pbe.Job

	// Helper which sets up Task.Resources struct to non-nil
	blankJob := func() *pbe.Job {
		return &pbe.Job{Task: &pbe.Task{Resources: &pbe.Resources{}}}
	}

	// test CPUs too big
	j = blankJob()
	j.Task.Resources.MinimumCpuCores = 2
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to worker without enough CPU resources")
	}

	// test RAM too big
	j = blankJob()
	j.Task.Resources.MinimumRamGb = 2.0
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to worker without enough RAM resources")
	}

	// test disk too big
	j = blankJob()
	j.Task.Resources.Volumes = []*pbe.Volume{
		{SizeGb: 2.0},
	}

	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to worker without enough Disk resources")
	}

	// test two volumes, basically check that they are
	// added together to get total size
	j = blankJob()
	j.Task.Resources.Volumes = []*pbe.Volume{
		{SizeGb: 1.0},
		{SizeGb: 0.1},
	}

	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to worker without enough Disk resources, 2 volumes")
	}

	// test zones don't match
	j = blankJob()
	j.Task.Resources.Zones = []string{"test-zone"}
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to worker out of zone")
	}

	// Now test a job that fits
	j = blankJob()
	j.Task.Resources.MinimumCpuCores = 1
	j.Task.Resources.MinimumRamGb = 1.0
	j.Task.Resources.Volumes = []*pbe.Volume{
		{SizeGb: 0.5},
		{SizeGb: 0.5},
	}
	j.Task.Resources.Zones = []string{"ok-zone", "not-ok-zone"}

	o = s.Schedule(j)
	if o == nil {
		t.Error("Didn't schedule job when resources fit")
	}
}
