package dumb

import (
	"log"
	"sync/atomic"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
)

// Scheduler extends the core Scheduler interface with additional methods for tracking
// a count of available workers.
//
// TODO this is a bad name and interface.
// EXPERIMENTAL: I (buchanae) wrote this while exploring the Scheduler interfaces
// and how they might be composable. This is probably incorrect and poorly named,
// and ResourcePool might be a better name.
type Scheduler interface {
	sched.Scheduler
	Available() int
	IncrementAvailable()
	DecrementAvailable()
}

// NewScheduler returns a new dumb Scheduler instance.
func NewScheduler(workers int) Scheduler {
	return &scheduler{int32(workers)}
}

type scheduler struct {
	// TODO in a smarter scheduler, "available" might be "pool", which would be
	//      something like a list of nodes, each with a description of its resources.
	//      This scheduler would handle the tricky part of matching a task to a node,
	//      but nodes, starting them, assigning work to them, updating the "pool", etc.
	//      would be handled elsewhere.
	available int32
}

// Available returns an integer describing how many worker slots are available.
// TODO in a smarter scheduler, these would be replaced by the "pool"
func (s *scheduler) Available() int {
	avail := atomic.LoadInt32(&s.available)
	return int(avail)
}

// IncrementAvailable increments the number of available worker slots.
func (s *scheduler) IncrementAvailable() {
	atomic.AddInt32(&s.available, 1)
}

// DecrementAvailable decrements the number of available worker slots.
func (s *scheduler) DecrementAvailable() {
	atomic.AddInt32(&s.available, -1)
}

// Schedule schedules a job and returns a corresponding Offer.
// TODO in a smarter scheduler, this would handle the tricky parts of scheduling:
//      matching a task to the best node
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Println("Running local scheduler")

	// Make an offer if the current resource count is less than the max.
	// This is just a dumb placeholder for a future scheduler.
	//
	// A better algorithm would rank the tasks, have a concept of binpacking,
	// and be able to assign a task to a specific, best-match node.
	// This backend does none of that...yet.
	avail := s.Available()
	log.Printf("Available: %d", avail)
	if avail == 0 {
		return sched.RejectedOffer("Pool is full")
	}
	w := sched.Worker{
		ID: sched.GenWorkerID(),
		Resources: sched.Resources{
			CPU:  1,
			RAM:  1.0,
			Disk: 10.0,
		},
	}
	return sched.NewOffer(j, w)
}
