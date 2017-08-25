package local

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	poolmock "github.com/ohsu-comp-bio/funnel/node/mocks"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	sched "github.com/ohsu-comp-bio/funnel/scheduler"
	. "github.com/stretchr/testify/mock"
	"testing"
)

func init() {
	logger.Configure(logger.DebugConfig())
}

func simpleNode() *pbs.Node {
	return &pbs.Node{
		// This ID MUST match the ID set in setup()
		// because the local scheduler is built to have only a single node
		Id: "test-node-id",
		Resources: &pbs.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
		Available: &pbs.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
		State: pbs.NodeState_ALIVE,
		Zone:  "ok-zone",
	}
}

func setup(nodes []*pbs.Node) (*poolmock.Client, *Backend) {
	conf := config.Config{}
	mc := new(poolmock.Client)

	// Mock in test nodes
	mc.On("ListNodes", Anything, Anything, Anything).Return(&pbs.ListNodesResponse{
		Nodes: nodes,
	}, nil)

	s := &Backend{
		conf,
		mc,
		"test-node-id",
	}
	return mc, s
}

func TestNoNodes(t *testing.T) {
	_, s := setup([]*pbs.Node{})
	j := &tes.Task{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Task scheduled on empty nodes")
	}
}

func TestSingleNode(t *testing.T) {
	_, s := setup([]*pbs.Node{
		simpleNode(),
	})

	j := &tes.Task{}
	o := s.Schedule(j)
	if o == nil {
		t.Error("Failed to schedule task on single node")
		return
	}
	if o.Node.Id != "test-node-id" {
		t.Error("Scheduled task on unexpected node")
	}
}

// Test that the scheduler ignores nodes it doesn't own.
func TestIgnoreOtherNodes(t *testing.T) {
	other := simpleNode()
	other.Id = "other-node"

	_, s := setup([]*pbs.Node{other})

	j := &tes.Task{}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to other node")
	}
}

// Test that scheduler ignores nodes without the "ALIVE" state
func TestIgnoreNonAliveNodes(t *testing.T) {
	j := &tes.Task{}

	for name, val := range pbs.NodeState_value {
		w := simpleNode()
		w.State = pbs.NodeState(val)
		_, s := setup([]*pbs.Node{w})
		o := s.Schedule(j)

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
	_, s := setup([]*pbs.Node{
		simpleNode(),
	})

	var o *sched.Offer
	var j *tes.Task

	// Helper which sets up Task.Resources struct to non-nil
	blankTask := func() *tes.Task {
		return &tes.Task{Resources: &tes.Resources{}}
	}

	// test CPUs too big
	j = blankTask()
	j.Resources.CpuCores = 2
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to node without enough CPU resources")
	}

	// test RAM too big
	j = blankTask()
	j.Resources.RamGb = 2.0
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to node without enough RAM resources")
	}

	// test disk too big
	j = blankTask()
	j.Resources.SizeGb = 2.0
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to node without enough DiskGb resources")
	}

	// test zones don't match
	j = blankTask()
	j.Resources.Zones = []string{"test-zone"}
	o = s.Schedule(j)
	if o != nil {
		t.Error("Scheduled task to node out of zone")
	}

	// Now test a task that fits
	j = blankTask()
	j.Resources.CpuCores = 1
	j.Resources.RamGb = 1.0
	j.Resources.SizeGb = 1.0
	j.Resources.Zones = []string{"ok-zone", "not-ok-zone"}
	o = s.Schedule(j)
	if o == nil {
		t.Error("Didn't schedule task when resources fit")
	}
}
