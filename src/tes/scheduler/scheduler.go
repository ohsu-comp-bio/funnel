package scheduler

import (
	"context"
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

type Scaler interface {
	StartWorker(*pbr.Worker) error
	ShouldStartWorker(*pbr.Worker) bool
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

// DefaultScheduleAlgorithm implements a simple scheduling algorithm
// that is (currently) common across a few scheduler backends.
// Given a job, list of workers, and weights, it returns the best Offer or nil.
func DefaultScheduleAlgorithm(j *pbe.Job, workers []*pbr.Worker, weights config.Weights) *Offer {

	offers := []*Offer{}
	for _, w := range workers {
		// Filter out workers that don't match the job request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !Match(w, j, DefaultPredicates) {
			continue
		}

		sc := DefaultScores(w, j)
		sc = sc.Weighted(weights)

		offer := NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching workers were found.
	if len(offers) == 0 {
		return nil
	}

	SortByAverageScore(offers)
	return offers[0]
}

// ScaleLoop calls Scale in a ticker loop. The separation of the two
// is mostly useful for testing.
func ScaleLoop(db *server.TaskBolt, s Scaler, conf config.Config) {
	ticker := time.NewTicker(conf.Worker.TrackerRate)

	for {
		<-ticker.C
		Scale(db, s)
	}
}

// Scale implements some common logic for allowing scheduler backends
// to poll the database, looking for workers that need to be started
// and shutdown.
func Scale(db *server.TaskBolt, s Scaler) {

	resp, err := db.GetWorkers(context.TODO(), &pbr.GetWorkersRequest{})
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
		w.State = pbr.WorkerState_Initializing
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
