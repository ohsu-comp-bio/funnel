package condor

import (
  scheduler "tes/scheduler"
)

type TaskPlan struct {
  Task Task
  StartWorker bool
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

type WorkPlan {
  TaskPlans []TaskPlan
}

type counts {
  maxWorkers int
  activeWorkers int
  workers int
  openSlots int
}

type backend struct {
  factory Factory
  counts counts
}

func NewBackend(schedAddr string, proxyAddr string) backend {
	sched, _ := scheduler.NewClient(schedAddr)
	client, _ := NewProxyClient(proxyAddr)
	factory := Factory{schedAddr, client}
  return backend{factory, counts{}}
}

func (b backend) Plan(req scheduler.WorkRequest) (plan WorkPlan) {
  c := b.counts

  // Cluster is full, give up.
  if b.isFull() {
    return
  }

  // Filter out tasks which we know we can't support
  tasks := []Task
  for _, t := range req.Tasks {
    if c.isSupported(t) {
      tasks := append(tasks, t)
    }
  }

  count := len(tasks)

  // No supported tasks, give up.
  if count == 0 {
    return
  }

  // Determine how many new workers the cluster should add.
  newWorkers := math.Min(c.maxNewWorkers, math.Max(0, count - c.openSlots))
  // Determine how many tasks we can support.
  avail := math.Max(count, count - (newWorkers + c.openSlots))

  // Blindly accept the first `avail` tasks.
  // A better algorithm would rank the tasks, have a concept of binpacking,
  // and be able to assign a task to a specific, best-match node.
  // This backend does none of that. Workers will pick up the first task
  // they find, not the task that was assigned to them.
  for i := 0; i < avail; i++ {
    p := TaskPlan{
      Task: tasks[i],
      StartWorker: i < newWorkers,
    }
    plan.TaskPlans = append(plan.TaskPlans, p)
  }
  return
}

func (b backend) Execute(w WorkPlan) {
  for _, p := range w.TaskPlans {
    if p.StartWorker {
      b.factory.StartWorker()
    }
  }
  // TODO backend needs to track the tasks it owns?
}

func (b backend) isSupported(t Task) bool {
  // Check resource requirements
  // - compute cpu, ram, etc
  // - storage
  // - available apps and engines
  return true
}

// Factory is responsible for starting workers by submitting jobs to HTCondor.
type Factory struct {
	SchedAddr string
	condor    *ProxyClient
}

func (f Factory) StartWorker() {
	ctx := context.Background()
	req := &pbc.StartWorkerRequest{f.SchedAddr}
	f.condor.StartWorker(ctx, req)
}
