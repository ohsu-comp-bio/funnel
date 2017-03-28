package scheduler

import (
	pbe "funnel/ga4gh"
	pbr "funnel/server/proto"
)

// Predicate is a function that checks whether a job fits a worker.
type Predicate func(*pbe.Job, *pbr.Worker) bool

// ResourcesFit determines whether a job fits a worker's resources.
func ResourcesFit(j *pbe.Job, w *pbr.Worker) bool {
	req := j.Task.GetResources()

	switch {
	case w.GetPreemptible() && !req.GetPreemptible():
		log.Debug("Fail preemptible")
		return false
	case w.GetAvailable().GetCpus() <= 0:
		log.Debug("Fail zero cpus")
		return false
	case w.GetAvailable().GetRam() <= 0.0:
		log.Debug("Fail zero ram")
		return false
	case w.GetAvailable().GetCpus() < req.GetMinimumCpuCores():
		log.Debug("Fail cpus")
		return false
	case w.GetAvailable().GetRam() < req.GetMinimumRamGb():
		log.Debug("Fail ram")
		return false
	}
	return true
}

// VolumesFit determines whether a job's volumes fit a worker
// by checking that the worker has enough disk space available.
func VolumesFit(j *pbe.Job, w *pbr.Worker) bool {
	req := j.Task.GetResources()
	vol := req.GetVolumes()

	// Total size (GB) of all requested volumes
	var tot float64
	for _, v := range vol {
		tot += v.GetSizeGb()
	}

	if tot == 0.0 {
		return true
	}

	f := tot <= w.GetAvailable().GetDisk()
	if !f {
		log.Debug("Failed volumes", "tot", tot, "avail", w.GetAvailable().GetDisk())
	}
	return f
}

// PortsFit determines whether a job's ports fit a worker
// by checking that the worker has the requested ports available.
func PortsFit(j *pbe.Job, w *pbr.Worker) bool {
	// Get the set of active ports on the worker
	active := map[int32]bool{}
	for _, p := range w.ActivePorts {
		active[p] = true
	}
	// Loop through the requested ports, fail if they are active.
	for _, d := range j.GetTask().GetDocker() {
		for _, p := range d.Ports {
			h := p.GetHost()
			if h == 0 {
				// "0" means "assign a random port, so skip checking this one.
				continue
			}
			if b := active[h]; b {
				return false
			}
		}
	}
	return true
}

// ZonesFit determines whether a job's zones fit a worker.
func ZonesFit(j *pbe.Job, w *pbr.Worker) bool {
	if w.Zone == "" {
		// Worker doesn't have a set zone, so don't bother checking.
		return true
	}

	if len(j.GetTask().GetResources().GetZones()) == 0 {
		// Request doesn't specify any zones, so don't bother checking.
		return true
	}

	for _, z := range j.GetTask().GetResources().GetZones() {
		if z == w.Zone {
			return true
		}
	}
	log.Debug("Failed zones")
	return false
}

// NotDead returns true if the worker state is not Dead or Gone.
func NotDead(j *pbe.Job, w *pbr.Worker) bool {
	return w.State != pbr.WorkerState_Dead && w.State != pbr.WorkerState_Gone
}

// WorkerHasTag returns a predicate function which returns true
// if the worker has the given tag (key in Metadata field).
func WorkerHasTag(tag string) Predicate {
	return func(j *pbe.Job, w *pbr.Worker) bool {
		_, ok := w.Metadata[tag]
		return ok
	}
}

// DefaultPredicates is a list of Predicate functions that check
// the whether a job fits a worker.
var DefaultPredicates = []Predicate{
	ResourcesFit,
	VolumesFit,
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

// Match checks whether a job fits a worker using the given Predicate list.
func Match(worker *pbr.Worker, job *pbe.Job, predicates []Predicate) bool {
	for _, pred := range predicates {
		if ok := pred(job, worker); !ok {
			return false
		}
	}
	return true
}
