package scheduler

import (
	"context"
	"log"
	pbe "tes/ga4gh"
	"time"
)

// MultiScheduler combines the Scheduler and Coordinator interfaces,
// with the goal of providing an interface that can schedule jobs to multiple
// scheduler backends, e.g. Google Cloud, AWS, on-premise, etc.
type MultiScheduler interface {
	Scheduler
	Coordinator
}

// NewMultiScheduler returns a new MultiScheduler instance that
// coordinates scheduling jobs on multiple backends.
func NewMultiScheduler() MultiScheduler {
	return &multisched{
		Coordinator: NewCoordinator(),
		// TODO configurable duration, or pass Context to Schedule
		timeout: time.Second * 2,
	}
}

type multisched struct {
	Coordinator
	timeout time.Duration
}

// Schedule schedules a job on multiple backends.
func (m *multisched) Schedule(job *pbe.Job) Offer {
	log.Println("Running multi-scheduler")

	var best Offer
	// TODO should Schedule get a Context arg?
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	offers := m.Broadcast(ctx, job)

	for o := range offers {
		if !o.Rejected() {
			if best == nil {
				// There is no best plan so far, so accept this one.
				best = o
			} else {
				better, worse := m.compare(best, o)
				best = better
				worse.Reject()
			}
		}
	}
	return best
}

func (m *multisched) compare(a Offer, b Offer) (Offer, Offer) {
	// Compare the submitted plan to the current best plan.
	// There can be only one!
	// TODO for now, just always accept the first one, because
	//      I don't want to write the complex scheduling logic yet.
	//      More interesting logic for evaluating plans goes here.
	//      e.g. which plan costs less? How do you define cost?
	return a, b
}
