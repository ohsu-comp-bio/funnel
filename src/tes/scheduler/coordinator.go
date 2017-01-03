package scheduler

import (
	"context"
	pbe "tes/ga4gh"
)

type Coordinator interface {
	// TODO could pass Subscription to the Coordinator as an arg, which would allow
	//      for different types of subscriptions (e.g. buffered)
	Subscribe() Subscription
	Unsubscribe(Subscription)
	Broadcast(context.Context, *pbe.Task) <-chan Offer
	Submit(Offer)
}

type taskchan chan *pbe.Task
type Subscription <-chan *pbe.Task

func NewCoordinator() Coordinator {
	return &coordinator{
		make(chan *broadcast),
		make(chan taskchan),
		make(chan string),
		make(chan Offer),
	}
}

type coordinator struct {
	broadcasts    chan *broadcast
	subscriptions chan taskchan
	cleanup       chan string
	offers        chan Offer
}

func (c *coordinator) Subscribe() Subscription {
	s := make(chan *pbe.Task)
	c.subscriptions <- s
	return s
}

func (c *coordinator) Unsubscribe(s Subscription) {
	// TODO
}

func (c *coordinator) Submit(o Offer) {
	c.offers <- o
}

func (c *coordinator) Broadcast(ctx context.Context, t *pbe.Task) <-chan Offer {
	ch := make(chan Offer)
	cancelctx, cancelfunc := context.WithCancel(ctx)
	b := &broadcast{t, ch, 0, cancelfunc}
	c.broadcasts <- b
	go func() {
		<-cancelctx.Done()
		c.cleanup <- t.TaskID
	}()
	return ch
}

func (c *coordinator) Run() {
	state := make(map[string]*broadcast)
	subs := make([]taskchan, 1)
	for {
		select {
		case s := <-c.subscriptions:
			subs = append(subs, s)
		case b := <-c.broadcasts:
			state[b.task.TaskID] = b
			for _, s := range subs {
				go c.send(b.task, s)
			}
		case o := <-c.offers:
			id := o.Task().TaskID
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

func (c *coordinator) send(t *pbe.Task, s taskchan) {
	s <- t
}

type broadcast struct {
	task   *pbe.Task
	ch     chan Offer
	count  int
	cancel context.CancelFunc
}
