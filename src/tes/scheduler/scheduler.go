package scheduler

import (
	uuid "github.com/nu7hatch/gouuid"
	"tes/config"
	pbe "tes/ga4gh"
	server "tes/server"
	pbr "tes/server/proto"
	"time"
)

// Scheduler is responsible for scheduling a job. It has a single method which
// is responsible for taking a Job and returning an Offer, or nil if there is
// no worker matching the job request. An Offer includes the ID of the offered
// worker.
//
// Offers include scores which describe how well the job fits the worker.
// Scores may describe a wide variety of metrics: resource usage, packing,
// startup time, cost, etc. Scores and weights are used to control the behavior
// of schedulers, and to combine offers from multiple schedulers.
type Scheduler interface {
	Schedule(*pbe.Job) *Offer
}

// Offer describes a worker offered by a scheduler for a job.
// The Scores describe how well the job fits this worker,
// which could be used by other a scheduler to pick the best offer.
type Offer struct {
	JobID  string
	Worker *pbr.Worker
	Scores Scores
}

// NewOffer returns a new Offer instance.
func NewOffer(w *pbr.Worker, j *pbe.Job, s Scores) *Offer {
	return &Offer{
		JobID:  j.JobID,
		Worker: w,
		Scores: s,
	}
}

// ScheduleLoop starts a scheduling loop, pulling conf.ScheduleChunk jobs from the database,
// and sending those to the given scheduler.
//
// TODO make a way to stop this loop
func ScheduleLoop(db *server.TaskBolt, sched Scheduler, conf config.Config) {
	tickChan := time.NewTicker(conf.ScheduleRate).C

	for {
		<-tickChan
		Schedule(db, sched, conf)
	}
}

// Schedule runs a single job scheduling tick.
func Schedule(db *server.TaskBolt, sched Scheduler, conf config.Config) {
	db.CheckWorkers()
	for _, job := range db.ReadQueue(conf.ScheduleChunk) {
		offer := sched.Schedule(job)
		if offer != nil {
			log.Debug("Assigning job to worker",
				"jobID", job.JobID,
				"workerID", offer.Worker.Id,
			)
			db.AssignJob(job, offer.Worker)
		} else {
			log.Info("No worker could be scheduled for job", "jobID", job.JobID)
		}
	}
}

// GenWorkerID returns a UUID string.
func GenWorkerID() string {
	u, _ := uuid.NewV4()
	return "worker-" + u.String()
}
