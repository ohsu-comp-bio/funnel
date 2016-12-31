package local

import (
  "log"
  "tes/scheduler"
  pbe "tes/ga4gh"
	uuid "github.com/nu7hatch/gouuid"
)

type plan struct {
  scheduler.TaskPlan
  startWorker bool
  factory *factory
}

func (p *plan) Execute() {
  if p.startWorker {
    p.factory.StartWorker(p.WorkerID())
  }
}


func NewBackend() scheduler.Backend {
	factory := &factory{}
  // TODO hard-coded
  return backend{factory, counts{maxWorkers: 10}}
}

type backend struct {
  factory *factory
  counts counts
}

func (b backend) Plan(wl scheduler.Workload) []scheduler.TaskPlan {
  log.Println("Planning workload for local backend")
  plans := make([]scheduler.TaskPlan, 0)
  c := b.counts

  if c.idleWorkers == 0 && c.maxWorkers == c.activeWorkers {
    // Cluster is full, give up.
    return plans
  }

  // Filter out tasks which we know we can't support
  tasks := make([]*pbe.Task, 0)
  for _, t := range wl {
    if b.isSupported(t) {
      tasks = append(tasks, t)
    }
  }

  count := len(tasks)

  // No supported tasks, give up.
  if count == 0 {
    return plans
  }

  // Determine how many new workers the cluster should add.
  maxNewWorkers := c.maxWorkers - (c.activeWorkers + c.idleWorkers)
  newWorkers := min(maxNewWorkers, max(0, count - c.idleWorkers))
  // Determine how many tasks we can support.
  avail := max(count, newWorkers + c.idleWorkers)

  // Blindly accept the first `avail` tasks.
  // A better algorithm would rank the tasks, have a concept of binpacking,
  // and be able to assign a task to a specific, best-match node.
  // This backend does none of that...yet.
  for i := 0; i < avail; i++ {
    workerID := GenWorkerId()
    tp := scheduler.NewPlan(tasks[i].TaskID, workerID)
    // TODO get rid of start worker for local backend. always start a worker.
    startWorker := i < newWorkers
    p := &plan{tp, startWorker, b.factory}
    plans = append(plans, p)
  }
  return plans
}

func GenWorkerId() string {
	u, _ := uuid.NewV4()
	return u.String()
}

func (b backend) isSupported(t *pbe.Task) bool {
  // Check resource requirements
  // - compute cpu, ram, etc
  // - storage
  // - available apps and engines
  return true
}

// TODO need some sort of bookkeeping of which workers are active
type counts struct {
  maxWorkers int
  activeWorkers int
  idleWorkers int
}

//...I cannot believe I have to define this.
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
