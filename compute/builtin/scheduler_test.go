package builtin

import (
	"context"
	"testing"
	"time"

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

// Test that scheduler ignores nodes without the "ALIVE" state
func TestIgnoreNonAliveNodes(t *testing.T) {
	conf := testConfig()
	s := newTestSched(conf)
	n := newTestNode(conf)
	n.detail.State = NodeState_DEAD

	n.Start()

	// Give the scheduler server time to start.
	time.Sleep(300 * time.Millisecond)
	err := s.scheduleOne(&tes.Task{Id: "task-1"})
	if !isNoOfferError(err) {
		t.Fatal("expected noOfferError")
	}
}

// Test whether the scheduler correctly filters nodes based on
// cpu, ram, disk, etc.
func TestMatch(t *testing.T) {
	conf := testConfig()
	conf.Node.Resources.Cpus = 1
	conf.Node.Resources.RamGb = 1.0
	conf.Node.Resources.DiskGb = 1.0
	conf.Node.Zone = "ok-zone"
	s := newTestSched(conf)
	n := newTestNode(conf)
	n.workerRun = func(context.Context, string) error { return nil }

	n.Start()

	// Give the scheduler server time to start.
	time.Sleep(300 * time.Millisecond)

	// test CPUs too big
	err := s.scheduleOne(&tes.Task{Id: "task-1", Resources: &tes.Resources{CpuCores: 2}})
	if !isNoOfferError(err) {
		t.Error("Scheduled task to node without enough CPU resources")
	}

	// test RAM too big
	err = s.scheduleOne(&tes.Task{Id: "task-2", Resources: &tes.Resources{RamGb: 2.0}})
	if !isNoOfferError(err) {
		t.Error("Scheduled task to node without enough RAM resources")
	}

	// test disk too big
	err = s.scheduleOne(&tes.Task{Id: "task-3", Resources: &tes.Resources{DiskGb: 2.0}})
	if !isNoOfferError(err) {
		t.Error("Scheduled task to node without enough DiskGb resources")
	}

	// test zones don't match
	err = s.scheduleOne(&tes.Task{
		Id:        "task-4",
		Resources: &tes.Resources{Zones: []string{"test-zone"}},
	})
	if !isNoOfferError(err) {
		t.Error("Scheduled task to node out of zone")
	}

	// Now test a task that fits
	err = s.scheduleOne(&tes.Task{
		Id: "task-5",
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1.0,
			DiskGb:   1.0,
			Zones:    []string{"ok-zone", "not-ok-zone"},
		},
	})
	if isNoOfferError(err) {
		t.Error("Didn't schedule task when resources fit")
	}

	time.Sleep(300 * time.Millisecond)
}

// Test that a task requesting nothing has a minimum CPU of 1.
func TestMinimumCpuRequest(t *testing.T) {
	conf := testConfig()
	conf.Node.Resources.Cpus = 2
	conf.Node.Resources.RamGb = 1.0
	conf.Node.Resources.DiskGb = 1.0
	s := newTestSched(conf)
	n := newTestNode(conf)
	n.workerRun = func(context.Context, string) error {
		time.Sleep(10 * time.Second)
		return nil
	}

	n.Start()
	time.Sleep(2 * time.Second)

	// test CPUs too big
	err := s.scheduleOne(&tes.Task{Id: "task-1"})
	if isNoOfferError(err) {
		t.Fatal("expected task to be scheduled")
	}

	n.ping()
	time.Sleep(2 * time.Second)

	if n.detail.Available.Cpus != 1 {
		t.Errorf("expected node to have 1 CPU available, but got %d", n.detail.Available.Cpus)
	}

	// The async scheduler/node ping communication needs some time to complete.
	time.Sleep(2 * time.Second)

	r := s.handles[n.detail.Id].node.Available
	if r.Cpus != 1 {
		t.Errorf("expected node handle to have 1 CPU available, but got %d", r.Cpus)
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

*/
