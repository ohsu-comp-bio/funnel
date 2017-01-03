package openstack

import (
	"log"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	dumb "tes/scheduler/dumb"
)

func NewScheduler(workers int, conf Config) sched.Scheduler {
	return &scheduler{dumb.NewScheduler(workers), conf}
}

type scheduler struct {
	ds   dumb.Scheduler
	conf Config
}

func (s *scheduler) Schedule(t *pbe.Task) sched.Offer {
	log.Println("Running dumb openstack scheduler")

	o := s.ds.Schedule(t)
	go s.observe(o)
	return o
}

func (s *scheduler) observe(o sched.Offer) {
	<-o.Wait()

	if o.Accepted() {
		s.ds.DecrementAvailable()
		start(o.Worker().ID, s.conf)
		// TODO there is nothing to watch the status of openstack workers yet,
		//      so this scheduler only does N jobs and then stops.

	} else if o.Rejected() {
		log.Println("Local offer was rejected")
	}
}
