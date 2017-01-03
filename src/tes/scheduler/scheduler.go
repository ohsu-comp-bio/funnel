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
  "time"
  "log"
  server "tes/server"
  pbe "tes/ga4gh"
)

type Scheduler interface {
  // TODO could be <-chan Offer, or Schedule(Workload, chan<- Offer)
  Schedule(*pbe.Task) Offer
}

func StartScheduling(db *server.TaskBolt, sched Scheduler) {
  ticker := time.NewTicker(time.Second * 5)
  defer ticker.Stop()
  for {
    <-ticker.C
    for _, t := range db.ReadQueue(10) {
      offer := sched.Schedule(t)
      if offer.Rejected() {
        log.Println("Rejected")
      } else {
        offer.Accept()
        db.AssignTask(offer.Task().TaskID, offer.Worker().ID)
      }
    }
  }
}
