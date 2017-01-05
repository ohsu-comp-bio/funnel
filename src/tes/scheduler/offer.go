package scheduler

import (
	pbe "tes/ga4gh"
)

// Offer represents an offer made by a scheduler to assign a job to a specific worker.
// An offer can be accepted or rejected. Offer.Wait() can be used to wait for the offer
// to be accepted/rejected.
//
// Here's a likely workflow of scheduling:
// - a simple loop pulls a queued job from the database and passes it to a scheduler.
// - the scheduler matches the job to a worker and returns an offer.
// - the scheduler also starts a goroutine to observe the offer by calling Offer.Wait()
// - the simple loop calls Offer.Accept()
// - the scheduler's goroutine observed the offer was accepted and starts a worker
//   for the job.
//
// It's also possible that the scheduler determined it didn't support the job,
// in which case the scheduler returns a rejected offer.
//
// Offers allow multiple systems to coordinate scheduling. For example, the multi-scheduler
// can accept multiple offers, pick the best one, accept it, and reject the rest.
type Offer interface {
  // Job returns the Job struct this offer relates to.
	Job() *pbe.Job
  // Worker returns the Worker struct this offer relates to.
  // The scheduler determines the worker to offer for this task.
	Worker() Worker
  // Accept accepts the offer. Accepting a rejected offer has no effect.
	Accept()
  // Reject rejects the offer. Rejecting an accepted offer has no effect.
	Reject()
  // Accepted returns true if the offer was accepted.
  // This is not the same as !Offer.Rejected() because the offer might be pending.
	Accepted() bool
  // Rejected returns true if the offer was rejected.
  // This is not the same as !Offer.Accepted() because the offer might be pending.
	Rejected() bool
  // RejectWithReason rejects the offer and give a reason why.
	RejectWithReason(string)
  // RejectionReason returns the reason the offer was rejected. Returns an empty
  // string if the offer wasn't rejected, or was rejected without a reason.
	RejectionReason() string
  // Wait waits for the Offer to be accepted/rejected. A goroutine can use this
  // to observe the result of an offer.
	Wait() <-chan struct{}
}

// Resources describe a set of computational resources, e.g. CPU, RAM, etc.
type Resources struct {
	CPU  int
	RAM  float32
	Disk float32
}

// Worker represents a worker node.
type Worker struct {
	ID string
	Resources Resources

	// TODO
	// In the future this could describe to the scheduler
	// the costs/benefits of running this job with this offer,
	// for example, it could convey that there is cached data
	// (data locality is desired).
	// Other ideas:
	// - hard/soft resource requirements. Can this backend offer
	//   higher performance?
	// - monetary cost: maybe this backend knows about the cost
	//   of the cloud instance.
	// - startup time: maybe this backend needs to start a worker
	//   and that could take a few minutes
	// - related job locality: maybe this job is online and related
	//   to others, and a shared, local, fast network would be better.
	// - SLA: maybe this backend can run this, but with the caveat that
	//   the job could be interrupted (AWS spot?). Or maybe this
	//   cluster is prone to having nodes go down.
}

// NewOffer returns a new Offer for the given Job + Worker.
func NewOffer(j *pbe.Job, w Worker) Offer {
	return &offer{
		job:    j,
		worker: w,
		done:   make(chan struct{}),
	}
}

// RejectedOffer returns a new rejected offer with the given reason.
// A scheduler might use this to signal that it can't support a job, for example.
func RejectedOffer(reason string) Offer {
	o := &offer{done: make(chan struct{})}
	o.RejectWithReason(reason)
	return o
}

type offer struct {
	job      *pbe.Job
	worker   Worker
	done     chan struct{}
	accepted bool
	rejected bool
	reason   string
}

func (o *offer) Job() *pbe.Job {
	return o.job
}

func (o *offer) Worker() Worker {
	return o.worker
}

func (o *offer) Accepted() bool {
	return o.accepted
}

func (o *offer) Rejected() bool {
	return o.rejected
}

func (o *offer) Accept() {
	select {
	case <-o.done:
    // If the offer is already accepted/rejected, don't do anything.
		return
	default:
		o.accepted = true
		close(o.done)
	}
}

func (o *offer) Reject() {
	select {
	case <-o.done:
    // If the offer is already accepted/rejected, don't do anything.
		return
	default:
		o.rejected = true
		close(o.done)
	}
}

func (o *offer) RejectWithReason(r string) {
	o.reason = r
	o.Reject()
}

func (o *offer) RejectionReason() string {
	return o.reason
}

func (o *offer) Wait() <-chan struct{} {
	return o.done
}
