package scheduler

import (
	"context"
	"log"
	pbe "tes/ga4gh"
	"time"
)

type SchedulerCoordinator interface {
	Scheduler
	Coordinator
}

func NewSchedulerCoordinator() SchedulerCoordinator {
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

func (m *multisched) Schedule(task *pbe.Task) Offer {
	log.Println("Running multi-scheduler")

	var best Offer
	// TODO should Schedule get a Context arg?
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	offers := m.Broadcast(ctx, task)

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
