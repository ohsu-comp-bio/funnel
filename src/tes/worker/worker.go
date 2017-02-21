package worker

import (
	"context"
	"errors"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	pbr "tes/server/proto"
	"time"
)

type logUpdateChan chan *pbr.UpdateJobLogsRequest

type Worker interface {
	Start()
}

type worker struct {
	conf config.Worker
	// Channel for job updates from job runners: stdout/err, ports, etc.
	logUpdates logUpdateChan
	sched      *schedClient
	log        logger.Logger
	resources  *pbr.Resources
}

func NewWorker(conf config.Worker) (Worker, error) {
	sched, err := newSchedClient(conf)
	if err != nil {
		return nil, err
	}

	log := logger.New("worker", "workerID", conf.ID)
	res := detectResources(conf.Resources)
	logUpdates := make(logUpdateChan)
	return &worker{conf, logUpdates, sched, log, res}, nil
}

// Run runs a worker with the given config. This is responsible for communication
// with the server and starting job runners
func (w *worker) Start() {
	w.log.Info("Starting worker")
	// TODO need a way to shut the worker down.
	go w.watchLogs()
	go w.watchJobs()
}

// TODO need a way to shut the worker down.
// TODO Close isn't correct. Wouldn't actually stop internal goroutines.
//func (w *worker) Close() {
// Tell the scheduler that the worker is gone.
//w.sched.WorkerGone()
//w.sched.Close()
//}

func (w *worker) watchLogs() {
	for {
		up := <-w.logUpdates
		// UpdateJobLogs() is more lightweight than UpdateWorker(),
		// which is why it happens separately and at a different rate.
		err := w.sched.UpdateJobLogs(up)
		if err != nil {
			// TODO if the request failed, the job update is lost and the logs
			//      are corrupted. Cache logs to prevent this?
			w.log.Error("Job log update failed", err)
		}
	}
}

func (w *worker) watchJobs() {
	// Tracks active job runners: job ID -> jobRunner instance
	runners := map[string]*jobRunner{}

	ticker := time.NewTicker(w.conf.UpdateRate)
	defer ticker.Stop()

	for {
		<-ticker.C
		r, _ := w.sched.GetWorker(context.Background(), &pbr.GetWorkerRequest{Id: w.conf.ID})

		// Reconcile server state with worker state.
		rerr := w.reconcile(runners, r.Jobs)
		if rerr != nil {
			// TODO what's the best behavior here?
			log.Error("Couldn't reconcile worker state.", rerr)
			continue
		}

		// Worker data has been updated. Send back to server for database update.
		r.LastPing = time.Now().Unix()
		r.Resources = w.resources
		r.State = pbr.WorkerState_Alive
		_, err := w.sched.UpdateWorker(r)
		if err != nil {
			log.Error("Couldn't save worker update. Recovering.", err)
		}
	}
}

// reconcile merges the server state with the worker state:
// - identifies new jobs and starts new runners for them
// - identifies canceled jobs and cancels existing runners
// - updates pbr.Job structs with current job state (running, complete, error, etc)
func (w *worker) reconcile(runners map[string]*jobRunner, jobs map[string]*pbr.JobWrapper) error {
	var (
		Unknown  = pbe.State_Unknown
		Canceled = pbe.State_Canceled
	)

	// Combine job IDs from response with job IDs from runners so we can reconcile
	// both sets below.
	jobIDs := []string{}
	for jobID := range runners {
		jobIDs = append(jobIDs, jobID)
	}
	for jobID := range jobs {
		jobIDs = append(jobIDs, jobID)
	}

	for _, jobID := range jobIDs {
		wrapper := jobs[jobID]
		job := wrapper.Job
		runner := runners[jobID]
		jobSt := job.GetState()
		runSt := runner.State()

		switch {
		case jobSt == runSt:
			// States match, do nothing.

		case isActive(jobSt) && runSt == Unknown:
			// Job needs to be started.
			r := newJobRunner(w.conf, wrapper, w.logUpdates)
			runners[jobID] = r
			go r.Run()

		case isActive(jobSt) && runSt != Unknown:
			// Job is running, update state.
			job.State = runSt
			if isComplete(runSt) {
				delete(runners, jobID)
			}

		case jobSt == Canceled:
			// Job is canceled.
			// runner.Cancel() is idempotent, so blindly cancel and delete.
			runner.Cancel()
			delete(runners, jobID)

		case jobSt == Unknown && runSt != Unknown:
			// Edge case. There's a runner for a non-existant job. Delete it.
			// TODO is it better to leave it? Continue in absence of explicit command principle?
			runner.Cancel()
			delete(runners, jobID)
		case isComplete(jobSt) && isActive(runSt):
			// Edge case. The job is complete but the runner is still running.
			// This shouldn't happen but it's better to check for it anyway.
			runner.Cancel()
			delete(runners, jobID)
		case isActive(jobSt) && runSt == Canceled:
			// Edge case. Server says running but runner says canceled.
			// Not sure how this would happen.
			fallthrough
		case isComplete(jobSt) && runSt == Unknown:
			// Edge case. Job is complete and there's no runner. Do nothing.
			fallthrough
		case isComplete(jobSt) && runSt == Canceled:
			// Edge case. Job is complete and the runner is canceled. Do nothing.
			// This shouldn't happen but it's better to check for it anyway.
			//
			// Log so that these unexpected cases can be explored.
			log.Error("Edge case during worker reconciliation. Recovering.",
				"job", job, "runner", runner)

		default:
			log.Error("Unhandled case during worker reconciliation.",
				"job", job, "runner", runner)
			return errors.New("Unhandled case during worker reconciliation")
		}
	}
	return nil
}

func isActive(s pbe.State) bool {
	return s == pbe.State_Queued || s == pbe.State_Initializing || s == pbe.State_Running
}

func isComplete(s pbe.State) bool {
	return s == pbe.State_Complete || s == pbe.State_Error || s == pbe.State_SystemError
}
