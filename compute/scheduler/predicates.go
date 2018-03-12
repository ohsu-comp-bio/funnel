package scheduler

import (
	"fmt"

	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Predicate is a function that checks whether a task fits a node.
type Predicate func(*tes.Task, *pbs.Node) error

// ResourcesFit determines whether a task fits a node's resources.
func ResourcesFit(t *tes.Task, n *pbs.Node) error {
	req := t.GetResources()

	switch {
	case n.GetAvailable().GetCpus() <= 0:
		return fmt.Errorf("Fail zero cpus available")
	case n.GetAvailable().GetRamGb() <= 0.0:
		return fmt.Errorf("Fail zero ram available")
	case n.GetAvailable().GetDiskGb() <= 0.0:
		return fmt.Errorf("Fail zero disk available")
	case n.GetAvailable().GetCpus() < req.GetCpuCores():
		return fmt.Errorf(
			"Fail cpus, requested %d, available %d",
			req.GetCpuCores(),
			n.GetAvailable().GetCpus(),
		)
	case n.GetAvailable().GetRamGb() < req.GetRamGb():
		return fmt.Errorf(
			"Fail ram, requested %f, available %f",
			req.GetRamGb(),
			n.GetAvailable().GetRamGb(),
		)
	case n.GetAvailable().GetDiskGb() < req.GetDiskGb():
		return fmt.Errorf(
			"Fail disk, requested %f, available %f",
			req.GetDiskGb(),
			n.GetAvailable().GetDiskGb(),
		)
	}
	return nil
}

// NotDead returns true if the node state is not Dead or Gone.
func NotDead(j *tes.Task, n *pbs.Node) error {
	if n.State != pbs.NodeState_DEAD && n.State != pbs.NodeState_GONE {
		return nil
	}
	return fmt.Errorf("Fail not dead")
}

// Alive returns true if the node state is not Dead or Gone.
func Alive(j *tes.Task, n *pbs.Node) error {
	if n.State != pbs.NodeState_ALIVE {
		return fmt.Errorf("Fail not alive")
	}
	return nil
}

// NodeHasTag returns a predicate function which returns true
// if the node has the given tag (key in Metadata field).
func NodeHasTag(tag string) Predicate {
	return func(j *tes.Task, n *pbs.Node) error {
		if _, ok := n.Metadata[tag]; !ok {
			return fmt.Errorf("fail node has tag: %s", tag)
		}
		return nil
	}
}

// TODO should have a predicate which understands authorization
//      - storage
//      - other auth resources?
//      - does storage need to be scheduler specific?
//      - how can we detect that a task cannot ever be scheduled? can we?
//        for example, if it requests access to storage that isn't available?
//        maybe set a max. time allowed to be unscheduled before notification

// Match checks whether a task fits a node using the given Predicate list.
func Match(node *pbs.Node, task *tes.Task, predicates []Predicate) bool {
	for _, pred := range predicates {
		if err := pred(task, node); err != nil {
			return false
		}
	}
	return true
}
