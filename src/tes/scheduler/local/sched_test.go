package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"tes/config"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	pbr "tes/server/proto"
	"testing"
)

func simpleWorker() *pbr.Worker {
	return &pbr.Worker{
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

type mockClient struct {
	Workers []*pbr.Worker
}

func (mc *mockClient) GetWorkers(ctx context.Context, req *pbr.GetWorkersRequest, opts ...grpc.CallOption) (*pbr.GetWorkersResponse, error) {
	resp := &pbr.GetWorkersResponse{Workers: mc.Workers}
	return resp, nil
}

func setup() (*mockClient, *scheduler) {
	conf := config.Config{}
	mc := &mockClient{}
	s := &scheduler{
		conf,
		mc,
		"test-worker-id",
	}
	return mc, s
}

func TestNoWorkers(t *testing.T) {
	_, s := setup()
	j := &pbe.Job{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Job scheduled on empty workers")
	}
}

func TestSingleWorker(t *testing.T) {
	mc, s := setup()
	mc.Workers = append(mc.Workers, simpleWorker())
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
	mc, s := setup()
	other := simpleWorker()
	other.Id = "other-worker"
	mc.Workers = append(mc.Workers, other)
	j := &pbe.Job{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Scheduled job to other worker")
	}
}

// Test that scheduler ignores workers without the "Alive" state
func TestIgnoreNonAliveWorkers(t *testing.T) {
	mc, s := setup()
	w := simpleWorker()
	j := &pbe.Job{}
	mc.Workers = append(mc.Workers, w)

	for name, val := range pbr.WorkerState_value {
		w.State = pbr.WorkerState(val)
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
	mc, s := setup()
	mc.Workers = append(mc.Workers, simpleWorker())
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
