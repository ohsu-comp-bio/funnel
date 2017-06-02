package scheduler

import (
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Predicate is a function that checks whether a task fits a worker.
type Predicate func(*tes.Task, *pbf.Worker) bool

// ResourcesFit determines whether a task fits a worker's resources.
func ResourcesFit(t *tes.Task, w *pbf.Worker) bool {
	req := t.GetResources()

	switch {
	case w.GetPreemptible() && !req.GetPreemptible():
		log.Debug("Fail preemptible")
		return false
	case w.GetAvailable().GetCpus() <= 0:
		log.Debug("Fail zero cpus available")
		return false
	case w.GetAvailable().GetRamGb() <= 0.0:
		log.Debug("Fail zero ram available")
		return false
	case w.GetAvailable().GetDiskGb() <= 0.0:
		log.Debug("Fail zero disk available")
		return false
	case w.GetAvailable().GetCpus() < req.GetCpuCores():
		log.Debug(
			"Fail cpus",
			"requested", req.GetCpuCores(),
			"available", w.GetAvailable().GetCpus(),
		)
		return false
	case w.GetAvailable().GetRamGb() < req.GetRamGb():
		log.Debug(
			"Fail ram",
			"requested", req.GetRamGb(),
			"available", w.GetAvailable().GetRamGb(),
		)
		return false
	case w.GetAvailable().GetDiskGb() < req.GetSizeGb():
		log.Debug(
			"Fail disk",
			"requested", req.GetSizeGb(),
			"available", w.GetAvailable().GetDiskGb(),
		)
		return false
	}
	return true
}

// PortsFit determines whether a task's ports fit a worker
// by checking that the worker has the requested ports available.
func PortsFit(t *tes.Task, w *pbf.Worker) bool {
	// Get the set of active ports on the worker
	active := map[int32]bool{}
	for _, p := range w.ActivePorts {
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

// ZonesFit determines whether a task's zones fit a worker.
func ZonesFit(t *tes.Task, w *pbf.Worker) bool {
	if w.Zone == "" {
		// Worker doesn't have a set zone, so don't bother checking.
		return true
	}

	if len(t.GetResources().GetZones()) == 0 {
		// Request doesn't specify any zones, so don't bother checking.
		return true
	}

	for _, z := range t.GetResources().GetZones() {
		if z == w.Zone {
			return true
		}
	}
	log.Debug("Failed zones")
	return false
}

// NotDead returns true if the worker state is not Dead or Gone.
func NotDead(j *tes.Task, w *pbf.Worker) bool {
	return w.State != pbf.WorkerState_DEAD && w.State != pbf.WorkerState_GONE
}

// WorkerHasTag returns a predicate function which returns true
// if the worker has the given tag (key in Metadata field).
func WorkerHasTag(tag string) Predicate {
	return func(j *tes.Task, w *pbf.Worker) bool {
		_, ok := w.Metadata[tag]
		return ok
	}
}

// DefaultPredicates is a list of Predicate functions that check
// the whether a task fits a worker.
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

// Match checks whether a task fits a worker using the given Predicate list.
func Match(worker *pbf.Worker, task *tes.Task, predicates []Predicate) bool {
	for _, pred := range predicates {
		if ok := pred(task, worker); !ok {
			return false
		}
	}
	return true
}
