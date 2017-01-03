package scheduler

import (
  "context"
  pbe "tes/ga4gh"
)

type Coordinator interface {
  Subscribe(Scheduler)
  Broadcast(context.Context, *pbe.Task) <-chan Offer
}

func NewCoordinator() Coordinator {
  return &coordinator{
    make(chan *broadcast),
    make(chan Scheduler),
    make(chan string),
  }
}

type broadcast struct {
  task *pbe.Task
  ch chan Offer
  count int
  cancel context.CancelFunc
}

type coordinator struct {
  broadcasts chan *broadcast
  subscriptions chan Scheduler
  cleanup chan string
}

// TODO Unsubscribe
func (c *coordinator) Subscribe(s Scheduler) {
  c.subscriptions <- s
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
  subs := make([]Scheduler, 1)
  offers := make(chan Offer)
  for {
    select {
    case s := <-c.subscriptions:
      subs = append(subs, s)
    case b := <-c.broadcasts:
      state[b.task.TaskID] = b
      for _, s := range subs {
        go c.schedule(b.task, s, offers)
      }
    case o := <-offers:
      id := o.Task().TaskID
      b := state[id]
      b.ch <- o
      b.count++
      // TODO is this good enough? There are probably weird edge cases with the timing
      //      of a [un]subscription + broadcast, but I hope that so rare that it doesn't
      //      matter
      if b.count >= len(subs) {
        b.cancel()
      }
    case id := <-c.cleanup:
      b := state[id]
      close(b.ch)
      delete(state, id)
    }
  }
}

func (c *coordinator) schedule(t *pbe.Task, s Scheduler, o chan Offer) {
  o <- s.Schedule(t)
}
