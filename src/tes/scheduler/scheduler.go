package scheduler

import (
  "time"
  "log"
  server "tes/server"
  pbe "tes/ga4gh"
	uuid "github.com/nu7hatch/gouuid"
)

type Scheduler interface {
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
        log.Printf("Rejected: %s", offer.RejectionReason())
      } else {
        offer.Accept()
        db.AssignTask(offer.Task().TaskID, offer.Worker().ID)
      }
    }
  }
}

func GenWorkerID() string {
	u, _ := uuid.NewV4()
	return u.String()
}
