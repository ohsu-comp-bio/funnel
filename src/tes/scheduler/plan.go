package scheduler

import (
	pbe "tes/ga4gh"
)

type Offer interface {
	Job() *pbe.Job
	Worker() Worker
	Accept()
	Reject()
	Accepted() bool
	Rejected() bool
	RejectWithReason(string)
	RejectionReason() string
	Wait() <-chan struct{}
}

type Resources struct {
	CPU  int
	RAM  float32
	Disk float32
}

type Worker struct {
	ID string
	Resources

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

func NewOffer(j *pbe.Job, w Worker) Offer {
	return &offer{
		job:    j,
		worker: w,
		done:   make(chan struct{}),
	}
}

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
		return
	default:
		o.accepted = true
		close(o.done)
	}
}

func (o *offer) Reject() {
	select {
	case <-o.done:
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
