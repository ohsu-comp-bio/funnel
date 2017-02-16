package scheduler

import (
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

type Predicate func(*pbe.Job, *pbr.Worker) bool

func ResourcesFit(j *pbe.Job, w *pbr.Worker) bool {
	req := j.Task.GetResources()

	// If the task didn't include resource requirements,
	// assume it fits.
	//
	// TODO think about whether this is the desired behavior
	if req == nil {
		return true
	}
	switch {
	case w.GetAvailable().GetCpus() < req.GetMinimumCpuCores():
		return false
	case w.GetAvailable().GetRam() < req.GetMinimumRamGb():
		return false
		// TODO check volumes
	}
	return true
}

var DefaultPredicates = []Predicate{
	ResourcesFit,
}

// TODO should have a predicate which understands authorization
//      - storage
//      - other auth resources?
//      - does storage need to be scheduler specific?
//      - how can we detect that a task cannot ever be scheduled? can we?
//        for example, if it requests access to storage that isn't available?

// TODO other predicate ideas
// - preemptible
// - zones
// - port checking
// - disk conflict
// - labels/selectors
// - host name

func Match(worker *pbr.Worker, job *pbe.Job, predicates []Predicate) bool {
	for _, pred := range predicates {
		if ok := pred(job, worker); !ok {
			return false
		}
	}
	return true
}
