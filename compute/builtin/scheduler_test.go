package builtin

import (
	"testing"

	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestScheduleZeroNodes(t *testing.T) {
  conf := testConfig()
	s := newTestSched(conf)
  err := s.scheduleOne(&tes.Task{Id: "task-1"})
  if err == nil {
    t.Fatal("expected no offer error")
  }
}

/*

func TestReadQueue(t *testing.T) {
	c := tests.DefaultConfig()
	c.Compute = "builtin"
	f := tests.NewFunnel(c)
	f.StartServer()

	for i := 0; i < 10; i++ {
		f.Run(`--sh 'echo 1'`)
	}
	time.Sleep(time.Second * 5)

	tasks := f.Scheduler.Queue.ReadQueue(10)

	if len(tasks) != 10 {
		t.Error("unexpected task count", len(tasks))
	}

	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	// test that read queue returns tasks in first in first out order
	for i := range tasks {
		j := min(i+1, len(tasks)-1)
		if tasks[i].CreationTime > tasks[j].CreationTime {
			t.Error("unexpected task sort order")
		}
	}
}

// Test that scheduler ignores nodes without the "ALIVE" state
func TestIgnoreNonAliveNodes(t *testing.T) {
	j := &tes.Task{}

	for name, val := range scheduler.NodeState_value {
		w := simpleNode()
		w.State = scheduler.NodeState(val)
		_, s := setup([]*scheduler.Node{w})
		o := s.GetOffer(j)

		if name == "ALIVE" {
			// Testing ALIVE just so I know this test is node as expected
			if o == nil {
				t.Error("Didn't schedule task to alive node")
			}
		} else {
			if o != nil {
				t.Errorf("Scheduled task to non-alive node: %s", name)
				return
			}
		}

	}
}

// Test whether the scheduler correctly filters nodes based on
// cpu, ram, disk, etc.
func TestMatch(t *testing.T) {
	_, s := setup([]*scheduler.Node{
		simpleNode(),
	})

	var o *scheduler.Offer
	var j *tes.Task

	// Helper which sets up Task.Resources struct to non-nil
	blankTask := func() *tes.Task {
		return &tes.Task{Resources: &tes.Resources{}}
	}

	// test CPUs too big
	j = blankTask()
	j.Resources.CpuCores = 2
	o = s.GetOffer(j)
	if o != nil {
		t.Error("Scheduled task to node without enough CPU resources")
	}

	// test RAM too big
	j = blankTask()
	j.Resources.RamGb = 2.0
	o = s.GetOffer(j)
	if o != nil {
		t.Error("Scheduled task to node without enough RAM resources")
	}

	// test disk too big
	j = blankTask()
	j.Resources.DiskGb = 2.0
	o = s.GetOffer(j)
	if o != nil {
		t.Error("Scheduled task to node without enough DiskGb resources")
	}

	// test zones don't match
	j = blankTask()
	j.Resources.Zones = []string{"test-zone"}
	o = s.GetOffer(j)
	if o != nil {
		t.Error("Scheduled task to node out of zone")
	}

	// Now test a task that fits
	j = blankTask()
	j.Resources.CpuCores = 1
	j.Resources.RamGb = 1.0
	j.Resources.DiskGb = 1.0
	j.Resources.Zones = []string{"ok-zone", "not-ok-zone"}
	o = s.GetOffer(j)
	if o == nil {
		t.Error("Didn't schedule task when resources fit")
	}
}
*/
