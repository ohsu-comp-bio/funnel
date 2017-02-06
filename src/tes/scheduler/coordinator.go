package scheduler

import (
	"context"
	pbe "tes/ga4gh"
)

// Coordinator is responsible for coordinating offers from multiple schedulers.
type Coordinator interface {
	// TODO could pass Subscription to the Coordinator as an arg, which would allow
	//      for different types of subscriptions (e.g. buffered)
	Subscribe() Subscription
	Unsubscribe(Subscription)
	Broadcast(context.Context, *pbe.Job) <-chan Offer
	Submit(Offer)
}

type jobchan chan *pbe.Job

// Subscription is a read-only channel of Jobs to be scheduled.
type Subscription <-chan *pbe.Job

// NewCoordinator returns a new Coordinator instance.
func NewCoordinator() Coordinator {
	return &coordinator{
		make(chan *broadcast),
		make(chan jobchan),
		make(chan string),
		make(chan Offer),
	}
}

type coordinator struct {
	broadcasts    chan *broadcast
	subscriptions chan jobchan
	cleanup       chan string
	offers        chan Offer
}

// Subscribe returns a new Subscription to this Coordinator.
func (c *coordinator) Subscribe() Subscription {
	s := make(chan *pbe.Job)
	c.subscriptions <- s
	return s
}

// Unsubscribe is currently unimplemented.
// It is meant to remove and clean up a Subscription.
func (c *coordinator) Unsubscribe(s Subscription) {
	// TODO
}

// Submit submits an Offer from schduler backend (subscriber) to this Coordinator.
func (c *coordinator) Submit(o Offer) {
	c.offers <- o
}

// Broadcast broadcasts a Job to all subscribers and returns a channel
// of Offers from subscribers for that Job.
func (c *coordinator) Broadcast(ctx context.Context, j *pbe.Job) <-chan Offer {
	ch := make(chan Offer)
	cancelctx, cancelfunc := context.WithCancel(ctx)
	b := &broadcast{j, ch, 0, cancelfunc}
	c.broadcasts <- b
	go func() {
		<-cancelctx.Done()
		c.cleanup <- j.JobID
	}()
	return ch
}

// Run runs the Coordinator, which starts a loop that communicates jobs and offers
// to/from subscribers.
func (c *coordinator) Run() {
	state := make(map[string]*broadcast)
	subs := make([]jobchan, 1)
	for {
		select {
		case s := <-c.subscriptions:
			subs = append(subs, s)
		case b := <-c.broadcasts:
			state[b.job.JobID] = b
			for _, s := range subs {
				go c.send(b.job, s)
			}
		case o := <-c.offers:
			id := o.Job().JobID
			b := state[id]
			if b != nil {
				b.ch <- o
				b.count++
				// TODO is this good enough? There are probably weird edge cases with the timing
				//      of a [un]subscription + broadcast, but I hope that so rare that it doesn't
				//      matter
				if b.count >= len(subs) {
					b.cancel()
				}
			}
		case id := <-c.cleanup:
			b := state[id]
			close(b.ch)
			delete(state, id)
		}
	}
}

func (c *coordinator) send(j *pbe.Job, jc jobchan) {
	jc <- j
}

type broadcast struct {
	job    *pbe.Job
	ch     chan Offer
	count  int
	cancel context.CancelFunc
}
