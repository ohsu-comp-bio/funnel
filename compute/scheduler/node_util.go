package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// SubtractResources subtracts the resources requested by "task" from
// the node resources "in".
func SubtractResources(t *tes.Task, in *pbs.Resources) *pbs.Resources {
	out := &pbs.Resources{
		Cpus:   in.GetCpus(),
		RamGb:  in.GetRamGb(),
		DiskGb: in.GetDiskGb(),
	}
	tres := t.GetResources()

	// Cpus are represented by an unsigned int, and if we blindly
	// subtract it will rollover to a very large number. So check first.
	rcpus := tres.GetCpuCores()
	if rcpus >= out.Cpus {
		out.Cpus = 0
	} else {
		out.Cpus -= rcpus
	}

	out.RamGb -= tres.GetRamGb()
	out.DiskGb -= tres.GetSizeGb()

	// Check minimum values.
	if out.Cpus < 0 {
		out.Cpus = 0
	}
	if out.RamGb < 0.0 {
		out.RamGb = 0.0
	}
	if out.DiskGb < 0.0 {
		out.DiskGb = 0.0
	}
	return out
}

// AvailableResources calculates available resources given a list of tasks
// and base resources.
//
// TODO include active ports. maybe move Available out of the protobuf message
//      and expect this helper to be used?
func AvailableResources(tasks []*tes.Task, res *pbs.Resources) *pbs.Resources {
	a := &pbs.Resources{
		Cpus:   res.GetCpus(),
		RamGb:  res.GetRamGb(),
		DiskGb: res.GetDiskGb(),
	}
	for _, t := range tasks {
		a = SubtractResources(t, a)
	}
	return a
}

// UpdateNodeState checks whether a node is dead/gone based on the last
// time it pinged.
func UpdateNodeState(nodes []*pbs.Node, conf config.Scheduler) []*pbs.Node {
	var updated []*pbs.Node
	for _, node := range nodes {
		prevState := node.State

		if node.State == pbs.NodeState_GONE {
			updated = append(updated, node)
			continue
		}

		if node.Version == 0 {
			// This shouldn't be happening, because nodes should be
			// created with LastPing, but give it the benefit of the doubt
			// and leave it alone.
			continue
		}

		lastPing := time.Unix(0, node.Version)
		d := time.Since(lastPing)

		if node.State == pbs.NodeState_UNINITIALIZED || node.State == pbs.NodeState_INITIALIZING {

			// The node is initializing, which has a more liberal timeout.
			if d > conf.NodeInitTimeout {
				// Looks like the node failed to initialize. Mark it dead
				node.State = pbs.NodeState_DEAD
			}

		} else if node.State == pbs.NodeState_DEAD && d > conf.NodeDeadTimeout {
			// The node has been dead for long enough.
			node.State = pbs.NodeState_GONE

		} else if d > conf.NodePingTimeout {
			// The node hasn't pinged in awhile, mark it dead.
			node.State = pbs.NodeState_DEAD
		}

		if prevState != node.State {
			updated = append(updated, node)
		}
	}
	return updated
}
