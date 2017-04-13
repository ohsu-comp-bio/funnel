package scheduler

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"golang.org/x/net/context"
	"os"
	"time"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	ReadQueue(n int) []*tes.Job
	AssignJob(*tes.Job, *pbf.Worker)
	CheckWorkers() error
	GetWorkers(context.Context, *pbf.GetWorkersRequest) (*pbf.GetWorkersResponse, error)
	UpdateWorker(context.Context, *pbf.Worker) (*pbf.UpdateWorkerResponse, error)
}

// NewScheduler returns a new Scheduler instance.
func NewScheduler(db Database, conf config.Config) (*Scheduler, error) {
	return &Scheduler{db, conf}, nil
}

// Scheduler handles scheduling tasks to workers and support many backends.
type Scheduler struct {
	db   Database
	conf config.Config
}

// Start starts the scheduling loops. This does not block.
// The scheduler will take a chunk of tasks from the queue,
// request the the configured backend schedule them, and
// act on offers made by the backend.
func (s *Scheduler) Start(ctx context.Context) error {
	var err error
	backend, err := NewBackend(s.conf.Scheduler, s.conf)
	if err != nil {
		return err
	}

	err = os.MkdirAll(s.conf.WorkDir, 0755)
	if err != nil {
		return err
	}

	// The scheduler and scaler loops are in separate goroutines
	// so that a long, blocking call in one doesn't block the other.

	// Start backend scheduler loop
	go func() {
		ticker := time.NewTicker(s.conf.ScheduleRate)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ScheduleChunk(ctx, s.db, backend, s.conf)
			}
		}
	}()

	// If the scheduler implements the Scaler interface,
	// start a scaler loop
	if b, ok := backend.(Scaler); ok {
		// Start backend scaler loop
		go func() {
			ticker := time.NewTicker(s.conf.ScheduleRate)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					Scale(ctx, s.db, b)
				}
			}
		}()
	}
	return nil
}

// ScheduleChunk does a scheduling iteration. It checks the health of workers
// in the database, gets a chunk of tasks from the queue (configurable by config.ScheduleChunk),
// and calls the given scheduler. If the scheduler returns a valid offer, the
// job is assigned to the offered worker.
func ScheduleChunk(ctx context.Context, db Database, b Backend, conf config.Config) {
	db.CheckWorkers()
	for _, job := range db.ReadQueue(conf.ScheduleChunk) {
		offer := b.Schedule(job)
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

// Scale implements some common logic for allowing scheduler backends
// to poll the database, looking for workers that need to be started
// and shutdown.
func Scale(ctx context.Context, db Database, s Scaler) {

	resp, err := db.GetWorkers(ctx, &pbf.GetWorkersRequest{})
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
		_, err := db.UpdateWorker(ctx, w)

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
