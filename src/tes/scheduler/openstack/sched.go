package openstack

import (
	"log"
	"tes"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	dumb "tes/scheduler/dumb"
)

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf tes.Config) sched.Scheduler {
	return &scheduler{
		dumb.NewScheduler(conf.Schedulers.Openstack.NumWorkers),
		conf,
	}
}

type scheduler struct {
	ds   dumb.Scheduler
	conf tes.Config
}

// Schedule schedules a job, returning an Offer.
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Println("Running dumb openstack scheduler")

	o := s.ds.Schedule(j)
	go s.observe(o)
	return o
}

func (s *scheduler) observe(o sched.Offer) {
	<-o.Wait()

	if o.Accepted() {
		s.ds.DecrementAvailable()
		s.start(o.Worker().ID)
		// TODO there is nothing to watch the status of openstack workers yet,
		//      so this scheduler only does N jobs and then stops.

	} else if o.Rejected() {
		log.Println("Local offer was rejected")
	}
}
