package scheduler

import (
  "log"
	"time"
)

type Coordinator interface {
  Workload() Workload
  Submit(TaskPlan) bool
}

func NewMultiScheduler(db *server.TaskBolt) Scheduler {
  return &multisched{
    sync: make(chan bool),
    plans: make(chan *planhelper),
  }
}

type multisched struct {
  sync chan bool
  plans chan *planhelper
  workload Workload
  children int
}

func (m *multisched) Workload() Workload {
  m.children++
  <-m.sync
  return m.workload
}

// Submit accepts a TaskPlan for evaluation by the scheduler and returns
// a state: Accepted, Rejected, etc.
func (m *multisched) Submit(tp TaskPlan) bool {
  // Wrap the TaskPlan in an internal struct which helps track evaluation status.
  p := &plan{TaskPlan: tp, response: make(chan bool, 1)}
  s.plans <- p
  // Blocks until the plan is evaluated
  return <-p.response
}

func (m *multisched) Schedule(wl Workload) []TaskPlan {
  log.Println("Starting schedule iteration")
  log.Printf("Len of workload: %d", len(workload))

  // Coordinating workload with the child schedulers requires some synchronization.
  // The `sync` channel is used to synchronize multiple concurrent schedulers
  // calling Workload().
  m.workload = wl
  close(m.sync)
  m.sync = make(chan bool)
  m.children = 0

  best := make(map[string]*planhelper)
  counts := make(map[string]int)
  valid := make(map[string]bool)

  for _, t := range wl {
    valid[t.TaskID] = true
  }

  // TODO configurable duration, or use context
  timer := time.NewTimer(time.Second * 2)

  loop:
  for {
    select {
    case p := <-s.plans:
      // It's possible that backends could submit plans for previous workloads,
      // so ensure the plan is part of this workload.
      if ok := valid[p.plan.TaskID()]; !ok {
        p.reject()
      } else {
        id := p.plan.TaskID()
        counts[id] += 1
        b := best[id]

        // The plan was rejected by the child scheduler.
        if p.plan.State() == Rejected {
          p.reject()
          break
        }

        if b == nil {
          // There is no best plan so far, so accept this one.
          best[id] = p
        } else {
          // Compare the submitted plan to the current best plan.
          // There can be only one!
          // TODO for now, just reject all after the first.
          //      more interesting logic for evaluating plans goes here.
          //      e.g. which plan costs less? How do you define cost?
          p.reject()
        }

        // We have received a plan from every child that we expected,
        // so don't wait until the timer is done, finalize the best plan now.
        if counts[id] == m.children {
          p.accept()
        }
        // TODO know when we've received all expected plans for all tasks
        //      and exit early
      }
    case <-timer.C:
      // The deadline has arrived. Finalize all the best plans.
      for _, b := range best {
        p.accept()
      }
      break loop
    }
  }
}


// plan helps track plans in the scheduler code below.
type planhelper struct {
  plan TaskPlan
  // response is used to communicate the result of the scheduler
  // evaluating the plan. This allows Submit() to block on this
  // channel while the scheduler asychronously evaluates the plan.
  response chan State
  // finalized helps ensure the plan is only finalized once.
  finalized bool
}

func (p *planhelper) accept() {
  if p.finalized { return }
  p.response <- true
  p.finalized = true
}

func (p *planhelper) reject() {
  if p.finalized { return }
  p.response <- false
  p.finalized = true
}
