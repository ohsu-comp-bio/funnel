package scheduler

// TODO this is the start of a very basic scaling algorithm
//      if the queue is bigger than N, add workers
//      Things to think about:
//      - how does scaling take resource requirements into consideration
//      - how do does scaler switch modes so that it's not checking scale
//        while new workers are starting?
//      - how does the scaling algorithm become reusable across multiple
//        backends?
//      - how does scaler detect that the condor queue is full?
//        i.e. there is no room to scale up.
//      - want to schedule a set of tasks completely in one go
//        to ensure that multiple schedulers are acting on the same set
//        of inputs in a coordinated fashion. Don't want data changing
//        in the middle of scheduling.

import (
  "log"
	"time"
  server "tes/server"
  pbe "tes/ga4gh"
)

type Workload []*pbe.Task

type State int
const (
  Accepted State = iota
  Rejected
  Proposed
)

type TaskPlan interface {
  TaskID() string
  WorkerID() string
  State() State
  SetState(State)
  Execute()
  // TODO
  // In the future this could describe to the scheduler
  // the costs/benefits of running this task with this plan,
  // for example, it could convey that there is cached data
  // (data locality is desired).
  // Other ideas:
  // - hard/soft resource requirements. Can this backend offer
  //   higher performance?
  // - monetary cost: maybe this backend knows about the cost
  //   of the cloud instance.
  // - startup time: maybe this backend needs to start a worker
  //   and that could take a few minutes
  // - related task locality: maybe this task is online and related
  //   to others, and a shared, local, fast network would be better.
  // - SLA: maybe this backend can run this, but with the caveat that
  //   the task could be interrupted (AWS spot?). Or maybe this
  //   cluster is prone to having nodes go down.
}

func NewPlan(id string, workerid string) TaskPlan {
  return &basePlan{id, workerid, Proposed}
}

type basePlan struct {
  id string
  workerID string
  state State
}
func (b *basePlan) TaskID() string {
  return b.id
}
func (b *basePlan) WorkerID() string {
  return b.workerID
}
func (b *basePlan) State() State {
  return b.state
}
func (b *basePlan) SetState(s State) {
  b.state = s
}
func (b *basePlan) Execute() {}

type Scheduler interface {
  Subscribe() Subscription
  SubmitPlan(TaskPlan) State
}

type Subscription interface {
  Workload() <-chan Workload
}

// plan helps track plans in the scheduler code below.
type plan struct {
  TaskPlan
  // response is used to communicate the result of the scheduler
  // evaluating the plan. This allows SubmitPlan to block on this
  // channel while the scheduler asychronously evaluates the plan.
  response chan State
  // finalized helps ensure the plan is only finalized once.
  finalized bool
}

// subscription helps track subscriptions in the scheduler code below.
type subscription struct {
  id int
  ch chan Workload
}

// Workload blocks until the scheduler broadcasts a new workload,
// at that point it returns a Workload instance.
// e.g.
//     for wl := range subscription.Workload() {
//       ...
//     }
func (s *subscription) Workload() <-chan Workload {
  return s.ch
}


func NewScheduler(db *server.TaskBolt) Scheduler {
  s := &scheduler{
    db: db,
    plans: make(chan *plan),
    subch: make(chan *subscription),
    unsubch: make(chan *subscription),
  }

  // Handle concurrent [un]subscribe calls
  // TODO move to Start()?
  go func() {
    subs := make(subscriptions)
    for {
      select {
      case sub := <-s.subch:
        log.Println("RECV")
        subs[sub.id] = sub
      case sub := <-s.unsubch:
        delete(subs, sub.id)
      default:
        s.schedule(subs)
      }
    }
  }()
  return s
}

type subscriptions map[int]*subscription

type scheduler struct {
  db *server.TaskBolt
  plans chan *plan
  subid int
  subch chan *subscription
  unsubch chan *subscription
}

func (s *scheduler) Subscribe() Subscription {
  log.Println("Subscribe()")
  s.subid++
  sub := &subscription{s.subid, make(chan Workload)}
  s.subch <- sub
  log.Println("SENT Subscribe()")
  return sub
}

func (s *scheduler) Unsubscribe(sub *subscription) {
  s.unsubch <- sub
}

// SubmitPlan accepts a TaskPlan for evaluation by the scheduler and returns
// a state: Accepted, Rejected, etc.
func (s *scheduler) SubmitPlan(tp TaskPlan) State {
  // Wrap the TaskPlan in an internal struct which helps track evaluation status.
  p := &plan{TaskPlan: tp, response: make(chan State)}
  s.plans <- p
  // Blocks until the plan is evaluated
  return <-p.response
}

// finalize finalizes a plan, marking the related task/job as active
// if appropriate, and returning the response to the backend
// (which is waiting on a response to SubmitPlan())
func (s *scheduler) finalize(p *plan) {
  if p.finalized {
    return
  }
  if p.State() == Accepted {
    s.db.AssignTask(p.TaskID(), p.WorkerID())
  }
  p.finalized = true
  p.response <- p.State()
}

// schedule executes a scheduling iteration. It is given a set of subscriptions,
// which are from scheduler backends, with which it will communicate. It will
// determine the workload it wants to schedule (a list of tasks/jobs), send
// those to the backends, accumlate plans from the backends, evaluate the
// best plans to accept, and reject others.
func (s *scheduler) schedule(subs subscriptions) {
  log.Println("Starting schedule iteration")
  log.Printf("Subscribers: %d", len(subs))
  // Get a workload for scheduling
  // TODO configurable workload size and other parameters
  workload := Workload(s.db.ReadQueue(10))

  log.Printf("Len of workload: %d", len(workload))

  // Send the workload to all the subscribers
  for _, sub := range subs {
    log.Println("Sending to sub")
    sub.ch <- workload
  }

  best := make(map[string]*plan)
  counts := make(map[string]int)
  valid := make(map[string]bool)

  for _, t := range workload {
    valid[t.TaskID] = true
  }

  // TODO configurable duration
  timer := time.NewTimer(time.Second * 2)

  for {
    select {
    case p := <-s.plans:
      // It's possible that backends could submit plans for previous workloads,
      // so ensure the plan is part of this workload.
      if ok := valid[p.TaskID()]; ok {
        id := p.TaskID()
        counts[id] += 1
        b := best[id]

        // The plan was rejected by the backend.
        if p.State() == Rejected {
          s.finalize(p)
          break
        }

        if b == nil {
          // There is no best plan so far, so accept this one.
          best[id] = p
          p.SetState(Accepted)
        } else {
          // Compare the submitted plan to the current best plan.
          // There can be only one!
          // TODO for now, just reject all after the first.
          //      more interesting logic for evaluating plans goes here.
          //      e.g. which plan costs less? How do you define cost?
          p.SetState(Rejected)
          s.finalize(p)
        }

        // We have received a plan from every backend that we expected,
        // so don't wait until the timer is done, finalize the best plan now.
        if counts[id] == len(subs) {
          s.finalize(p)
        }
      }
    case <-timer.C:
      // The deadline has arrived. Finalize all the best plans.
      for _, b := range best {
        s.finalize(b)
      }
      return
    }
  }
}


/*
func DumbLocalAutoscaler(schedAddr string) Autoscaler {
	sched, _ := scheduler.NewClient(schedAddr)
	factory := LocalWorkerFactory{}
	return &DumbAutoscaler{sched, factory}
}

func PollAutoscaler(autoscaler Autoscaler, d time.Duration) {
	ticker := time.NewTicker(d)
	for _ = range ticker.C {
		autoscaler.Analyze()
	}
}
*/
