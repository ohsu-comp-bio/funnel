package scheduler

import (
	uuid "github.com/nu7hatch/gouuid"
	pbe "tes/ga4gh"
	server "tes/server"
	"time"
)

// Scheduler is responsible for scheduling a job. It has a single method which
// is responsible for taking a job and returning an Offer which describes whether
// a scheduler can run the job, how many resources it can offer, and anything that
// might allow a central scheduler to decide where best to run the job.
//
// For example, a system might have a separate scheduler for each of
// Google Cloud, AWS, and on-premise HTCondor clusters. For a given job,
// the Google Cloud and AWS schedulers might determine that the job cannot be run
// (maybe due to data locality restrictions) and they return rejected Offers, while
// the HTCondor returns an accepted Offer. A central scheduler can then determine
// that the job should be assigned to the HTCondor cluster.
type Scheduler interface {
	Schedule(*pbe.Job) Offer
}

// StartScheduling starts a scheduling loop, pulling 10 jobs from the database,
// and sending those to the given scheduler. This happens every 5 seconds.
//
// Offers which aren't marked as rejected by the scheduler are accepted
// and the job is assigned to a worker in the database.
func StartScheduling(db *server.TaskBolt, sched Scheduler) {
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		<-ticker.C
		for _, t := range db.ReadQueue(10) {
			offer := sched.Schedule(t)
			if offer.Rejected() {
				log.Debug("Rejected offer", "reason", offer.RejectionReason())
			} else {
				log.Debug("Assigning job to worker",
					"jobID", offer.Job().JobID,
					"workerID", offer.Worker().ID,
				)
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
