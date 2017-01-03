package scheduler

import (
	uuid "github.com/nu7hatch/gouuid"
	"log"
	pbe "tes/ga4gh"
	server "tes/server"
	"time"
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
