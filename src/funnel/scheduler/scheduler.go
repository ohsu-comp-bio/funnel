package scheduler

import (
	"context"
	"fmt"
	"funnel/config"
	tes "funnel/proto/tes"
	server "funnel/server"
	pbf "funnel/proto/funnel"
	uuid "github.com/nu7hatch/gouuid"
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
	Schedule(*tes.Job) *Offer
}

// Scaler represents a service that can start worker instances, for example
// the Google Cloud Scheduler backend.
type Scaler interface {
	// StartWorker is where the work is done to start a worker instance,
	// for example, calling out to Google Cloud APIs.
	StartWorker(*pbf.Worker) error
	// ShouldStartWorker allows scalers to filter out workers they are interested in.
	// If "true" is returned, Scaler.StartWorker() will be called with this worker.
	ShouldStartWorker(*pbf.Worker) bool
}

// Offer describes a worker offered by a scheduler for a job.
// The Scores describe how well the job fits this worker,
// which could be used by other a scheduler to pick the best offer.
type Offer struct {
	JobID  string
	Worker *pbf.Worker
	Scores Scores
}

// NewOffer returns a new Offer instance.
func NewOffer(w *pbf.Worker, j *tes.Job, s Scores) *Offer {
	return &Offer{
		JobID:  j.JobID,
		Worker: w,
		Scores: s,
	}
}

// ScheduleChunk does a scheduling iteration. It checks the health of workers
// in the database, gets a chunk of tasks from the queue (configurable by config.ScheduleChunk),
// and calls the given scheduler. If the scheduler returns a valid offer, the
// job is assigned to the offered worker.
func ScheduleChunk(db server.Database, sched Scheduler, conf config.Config) {
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
func GenWorkerID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-worker-%s", prefix, u.String())
}

// Scale implements some common logic for allowing scheduler backends
// to poll the database, looking for workers that need to be started
// and shutdown.
func Scale(db server.Database, s Scaler) {

	resp, err := db.GetWorkers(context.TODO(), &pbf.GetWorkersRequest{})
	if err != nil {
		log.Error("Failed GetWorkers request. Recovering.", err)
		return
	}

	for _, w := range resp.Workers {

		if !s.ShouldStartWorker(w) {
			continue
		}

		serr := s.StartWorker(w)
		if serr != nil {
			log.Error("Error starting worker", serr)
			continue
		}

		// TODO should the Scaler instance handle this? Is it possible
		//      that Initializing is the wrong state in some cases?
		w.State = pbf.WorkerState_Initializing
		_, err := db.UpdateWorker(context.TODO(), w)

		if err != nil {
			// TODO an error here would likely result in multiple workers
			//      being started unintentionally. Not sure what the best
			//      solution is. Possibly a list of failed workers.
			//
			//      If the scheduler and database API live on the same server,
			//      this *should* be very unlikely.
			log.Error("Error marking worker as initializing", err)
		}
	}
}

// ScheduleLoop calls ScheduleChunk every config.ScheduleRate, in a loop.
//
// TODO make a way to stop this loop
func ScheduleLoop(db server.Database, sched Scheduler, conf config.Config) {
	tickChan := time.NewTicker(conf.ScheduleRate).C

	for {
		<-tickChan
		ScheduleChunk(db, sched, conf)
	}
}

// ScaleLoop calls Scale every config.ScheduleRate, in a loop
//
// TODO make a way to stop this loop
func ScaleLoop(db server.Database, s Scaler, conf config.Config) {
	ticker := time.NewTicker(conf.ScheduleRate)

	for {
		<-ticker.C
		Scale(db, s)
	}
}
