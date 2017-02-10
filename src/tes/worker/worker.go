package worker

import (
	"context"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	"tes/scheduler"
	pbr "tes/server/proto"
	"tes/util"
	"time"
)

type updateChan chan *pbr.UpdateStatusRequest

type result struct {
	Job *pbe.Job
	Err error
}

type Worker struct {
	ID      string
	conf    config.Worker
	log     logger.Logger
	jobs    chan *pbr.JobResponse
	sched   *scheduler.Client
	updates updateChan
}

func NewWorker(conf config.Worker) (*Worker, error) {
	sched, err := scheduler.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return &Worker{
		ID:      conf.ID,
		conf:    conf,
		log:     logger.New("worker", "workerID", conf.ID),
		sched:   sched,
		jobs:    make(chan *pbr.JobResponse),
		updates: make(updateChan),
	}, nil
}

// Close cleans up worker resources.
func (w *Worker) Close() {
	w.sched.Close()
}

func (w *Worker) Run(pctx context.Context) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	w.log.Info("Running worker")
	defer w.log.Info("Shutting down")

	go w.sched.PollForJobs(ctx, w.ID, w.jobs)
	w.trackJobs(ctx)
}

func (w *Worker) trackJobs(ctx context.Context) {
	runningJobs := 0
	done := make(chan bool)
	// timeout controls how long before the worker shuts down
	// when no jobs are available.
	timeout := util.NewIdleTimeout(w.conf.Timeout)

	// If no jobs are found for awhile, the worker will shut down.
	for {
		select {
		case <-ctx.Done():
			return

		// "w.updates" contains job updates, e.g. stdout/err updates.
		case up := <-w.updates:
			w.sched.UpdateJobStatus(ctx, up)

		case <-timeout.Done():
			w.log.Info("Reached idle timeout")
			return

		// "jobs" is written to when a new job is assigned by the scheduler.
		case job := <-w.jobs:
			log.Debug("Worker got job", "job", job)
			go func() {
				w.runJob(ctx, job)
				done <- true
			}()
			runningJobs++
			timeout.Stop()

			// "done" is written to when a job finishes
		case <-done:
			runningJobs--
			if runningJobs == 0 {
				timeout.Start()
			}
		}
	}
}

// runJob handles creating and calling the job runner and communicating
// job state with the scheduler, including watching for cancelation signals.
func (w *Worker) runJob(pctx context.Context, resp *pbr.JobResponse) {
	job := resp.Job
	ctx := w.jobContext(pctx, job.JobID)
	err := runJob(ctx, resp, w.conf, w.updates)
	failed := err != nil
	w.sched.JobComplete(ctx, &pbr.JobCompleteRequest{
		Id:     job.JobID,
		Failed: failed,
	})
}

// jobContext returns a context object that will be canceled when the job
// is canceled by the scheduler.
func (w *Worker) jobContext(parent context.Context, jobID string) context.Context {
	ctx, cancel := context.WithCancel(parent)

	go func() {
		ticker := time.NewTicker(w.conf.StatusPollRate)
		defer ticker.Stop()

		for {
			select {
			case <-parent.Done():
				return

			case <-ticker.C:
				resp, err := w.sched.GetJobState(ctx, &pbe.JobID{jobID})
				if err != nil {
					w.log.Error("Couldn't get job state. Recovering.", err)
				} else if resp.State == pbe.State_Canceled {
					cancel()
					return
				}
			}
		}
	}()
	return ctx
}
