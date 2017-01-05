package scheduler

import (
	uuid "github.com/nu7hatch/gouuid"
	"log"
	pbe "tes/ga4gh"
	server "tes/server"
	"time"
)

type Scheduler interface {
	Schedule(*pbe.Job) Offer
}

// StartScheduling starts a scheduling loop, pulling 10 jobs from the database,
// and sending those to the given scheduler. This happens every 5 seconds.
//
// Offers which aren't marked as rejected by the scheduler are accepted
// and the job is assigned to a worker in the database.
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
				db.AssignJob(offer.Job().JobID, offer.Worker().ID)
			}
		}
	}
}

// GenWorkerID returns a UUID string.
func GenWorkerID() string {
	u, _ := uuid.NewV4()
	return u.String()
}
