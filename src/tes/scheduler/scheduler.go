package scheduler

// TODO this is the start of a very basic scaling algorithm
//      if the queue is bigger than N, add workers
//      Things to think about:
//      - how does scaling take resource requirements into consideration
//      - how does the scaling algorithm become reusable across multiple
//        backends?
//      - how does scaler detect that the condor queue is full?
//        i.e. there is no room to scale up.

import (
  server "tes/server"
  pbe "tes/ga4gh"
)

type Workload []*pbe.Task

type Scheduler interface {
  Schedule(Workload) []TaskPlan
}

func StartScheduling(db *server.TaskBolt, s Scheduler) {
  workload := Workload(db.ReadQueue(10))
  plans := s.Schedule(ctx, workload)
  if p.State() == Accepted {
    s.db.AssignTask(p.TaskID(), p.WorkerID())
  }
}
