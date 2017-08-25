package scheduler

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Predicate is a function that checks whether a task fits a node.
type Predicate func(*tes.Task, *pbs.Node) bool

// ResourcesFit determines whether a task fits a node's resources.
func ResourcesFit(t *tes.Task, n *pbs.Node) bool {
	req := t.GetResources()

	switch {
	case n.GetPreemptible() && !req.GetPreemptible():
		log.Debug("Fail preemptible")
		return false
	case n.GetAvailable().GetCpus() <= 0:
		log.Debug("Fail zero cpus available")
		return false
	case n.GetAvailable().GetRamGb() <= 0.0:
		log.Debug("Fail zero ram available")
		return false
	case n.GetAvailable().GetDiskGb() <= 0.0:
		log.Debug("Fail zero disk available")
		return false
	case n.GetAvailable().GetCpus() < req.GetCpuCores():
		log.Debug(
			"Fail cpus",
			"requested", req.GetCpuCores(),
			"available", n.GetAvailable().GetCpus(),
		)
		return false
	case n.GetAvailable().GetRamGb() < req.GetRamGb():
		log.Debug(
			"Fail ram",
			"requested", req.GetRamGb(),
			"available", n.GetAvailable().GetRamGb(),
		)
		return false
	case n.GetAvailable().GetDiskGb() < req.GetSizeGb():
		log.Debug(
			"Fail disk",
			"requested", req.GetSizeGb(),
			"available", n.GetAvailable().GetDiskGb(),
		)
		return false
	}
	return true
}

// PortsFit determines whether a task's ports fit a node
// by checking that the node has the requested ports available.
func PortsFit(t *tes.Task, n *pbs.Node) bool {
	// Get the set of active ports on the node
	active := map[int32]bool{}
	for _, p := range n.ActivePorts {
		active[p] = true
	}
	// Loop through the requested ports, fail if they are active.
	for _, d := range t.GetExecutors() {
		for _, p := range d.Ports {
			h := p.GetHost()
			if h == 0 {
				// "0" means "assign a random port, so skip checking this one.
				continue
			}
			if b := active[int32(h)]; b {
				return false
			}
		}
	}
	return true
}

// ZonesFit determines whether a task's zones fit a node.
func ZonesFit(t *tes.Task, n *pbs.Node) bool {
	if n.Zone == "" {
		// Node doesn't have a set zone, so don't bother checking.
		return true
	}

	if len(t.GetResources().GetZones()) == 0 {
		// Request doesn't specify any zones, so don't bother checking.
		return true
	}

	for _, z := range t.GetResources().GetZones() {
		if z == n.Zone {
			return true
		}
	}
	log.Debug("Failed zones")
	return false
}

// NotDead returns true if the node state is not Dead or Gone.
func NotDead(j *tes.Task, n *pbs.Node) bool {
	return n.State != pbs.NodeState_DEAD && n.State != pbs.NodeState_GONE
}

// NodeHasTag returns a predicate function which returns true
// if the node has the given tag (key in Metadata field).
func NodeHasTag(tag string) Predicate {
	return func(j *tes.Task, n *pbs.Node) bool {
		_, ok := n.Metadata[tag]
		return ok
	}
}

// DefaultPredicates is a list of Predicate functions that check
// the whether a task fits a node.
var DefaultPredicates = []Predicate{
	ResourcesFit,
	PortsFit,
	ZonesFit,
	NotDead,
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
		if ok := pred(task, node); !ok {
			return false
		}
	}
	return true
}
