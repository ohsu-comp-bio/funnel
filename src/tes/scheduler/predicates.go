package scheduler

import (
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

type Predicate func(*pbe.Job, *pbr.Worker) bool

func ResourcesFit(j *pbe.Job, w *pbr.Worker) bool {
	req := j.Task.GetResources()

	switch {
	case w.GetPreemptible() && !req.GetPreemptible():
		return false
	case w.GetAvailable().GetCpus() <= 0:
		return false
	case w.GetAvailable().GetRam() <= 0.0:
		return false
	case w.GetAvailable().GetCpus() < req.GetMinimumCpuCores():
		return false
	case w.GetAvailable().GetRam() < req.GetMinimumRamGb():
		return false
		// TODO check volumes
	}
	return true
}

func VolumesFit(j *pbe.Job, w *pbr.Worker) bool {
	req := j.Task.GetResources()
	vol := req.GetVolumes()

	// Total size (GB) of all requested volumes
	var tot float64
	for _, v := range vol {
		tot += v.GetSizeGb()
	}

	return tot < w.GetAvailable().GetDisk()
}

func PortsFit(j *pbe.Job, w *pbr.Worker) bool {
	// Get the set of active ports on the worker
	active := map[int32]bool{}
	for _, p := range w.ActivePorts {
		active[p] = true
	}
	// Loop through the requested ports, fail if they are active.
	for _, d := range j.Task.Docker {
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

func ZonesFit(j *pbe.Job, w *pbr.Worker) bool {
	if w.Zone == "" {
		// Worker doesn't have a set zone, so don't bother checking.
		return true
	}

	for _, z := range j.Task.GetResources().Zones {
		if z == w.Zone {
			return true
		}
	}
	return false
}

var DefaultPredicates = []Predicate{
	ResourcesFit,
	VolumesFit,
	PortsFit,
	ZonesFit,
}

// TODO should have a predicate which understands authorization
//      - storage
//      - other auth resources?
//      - does storage need to be scheduler specific?
//      - how can we detect that a task cannot ever be scheduled? can we?
//        for example, if it requests access to storage that isn't available?
//        maybe set a max. time allowed to be unscheduled before notification

func Match(worker *pbr.Worker, job *pbe.Job, predicates []Predicate) bool {
	for _, pred := range predicates {
		if ok := pred(job, worker); !ok {
			return false
		}
	}
	return true
}
