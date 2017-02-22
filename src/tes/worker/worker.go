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

type runJobFunc func(JobControl, config.Worker, *pbr.JobWrapper, logUpdateChan)
type logUpdateChan chan *pbr.UpdateJobLogsRequest

// Worker represents a worker which processes jobs
// TODO better docs or just remove this interface?
type Worker interface {
	Run()
}

// NewWorker returns a new Worker instance
func NewWorker(conf config.Worker) (Worker, error) {
	sched, err := newSchedClient(conf)
	if err != nil {
		return nil, err
	}

	log := logger.New("worker", "workerID", conf.ID)
	res := detectResources(conf.Resources)
	logUpdates := make(logUpdateChan)
	// Tracks active job ctrls: job ID -> JobControl instance
	ctrls := map[string]JobControl{}
	return &worker{conf, logUpdates, sched, log, res, runJob, ctrls}, nil
}

type worker struct {
	conf config.Worker
	// Channel for job updates from job runners: stdout/err, ports, etc.
	logUpdates logUpdateChan
	sched      *schedClient
	log        logger.Logger
	resources  *pbr.Resources
	// runJob is here so it can be mocked during testing
	runJob runJobFunc
	ctrls  map[string]JobControl
}

// Run runs a worker with the given config. This is responsible for communication
// with the server and starting job runners
func (w *worker) Run() {
	w.log.Info("Starting worker")
	// TODO need a way to shut the worker down.

	ticker := time.NewTicker(w.conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		case up := <-w.logUpdates:
			w.updateLogs(up)
		case <-ticker.C:
			w.checkJobs()
		}
	}
}

// TODO need a way to shut the worker down.
// TODO Close isn't correct. Wouldn't actually stop internal goroutines.
//func (w *worker) Close() {
// Tell the scheduler that the worker is gone.
//w.sched.WorkerGone()
//w.sched.Close()
//}

func (w *worker) updateLogs(up *pbr.UpdateJobLogsRequest) {
	// UpdateJobLogs() is more lightweight than UpdateWorker(),
	// which is why it happens separately and at a different rate.
	err := w.sched.UpdateJobLogs(up)
	if err != nil {
		// TODO if the request failed, the job update is lost and the logs
		//      are corrupted. Cache logs to prevent this?
		w.log.Error("Job log update failed", err)
	}
}

func (w *worker) checkJobs() {
	r, _ := w.sched.GetWorker(context.TODO(), &pbr.GetWorkerRequest{Id: w.conf.ID})

	// Reconcile server state with worker state.
	rerr := w.reconcile(r.Jobs)
	if rerr != nil {
		// TODO what's the best behavior here?
		log.Error("Couldn't reconcile worker state.", rerr)
		return
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

// TODO this is trying to re-run jobs on failure
// reconcile merges the server state with the worker state:
// - identifies new jobs and starts new runners for them
// - identifies canceled jobs and cancels existing runners
// - updates pbr.Job structs with current job state (running, complete, error, etc)
func (w *worker) reconcile(jobs map[string]*pbr.JobWrapper) error {
	var (
		Unknown  = pbe.State_Unknown
		Canceled = pbe.State_Canceled
	)

	// Combine job IDs from response with job IDs from ctrls so we can reconcile
	// both sets below.
	jobIDs := map[string]bool{}
	for jobID := range w.ctrls {
		jobIDs[jobID] = true
	}
	for jobID := range jobs {
		jobIDs[jobID] = true
	}

	for jobID := range jobIDs {
		wrapper := jobs[jobID]
		job := wrapper.Job
		ctrl := w.ctrls[jobID]
		jobSt := job.GetState()
		runSt := pbe.State_Unknown
		if ctrl != nil {
			runSt = ctrl.State()
		}

		switch {
		case jobSt == Unknown && runSt == Unknown:
			// Edge case. Shouldn't be here, but log just in case.
			fallthrough
		case isActive(jobSt) && runSt == Canceled:
			// Edge case. Server says running but ctrl says canceled.
			// Not sure how this would happen.
			fallthrough
		case isComplete(jobSt) && runSt == Unknown:
			// Edge case. Job is complete and there's no ctrl. Do nothing.
			fallthrough
		case isComplete(jobSt) && runSt == Canceled:
			// Edge case. Job is complete and the ctrl is canceled. Do nothing.
			// This shouldn't happen but it's better to check for it anyway.
			//
			// Log so that these unexpected cases can be explored.
			log.Error("Edge case during worker reconciliation. Recovering.",
				"jobst", jobSt, "runst", runSt)

		case jobSt == runSt:
			// States match, do nothing.

		case isActive(jobSt) && runSt == Unknown:
			// Job needs to be started.
			ctrl := NewJobControl()
			w.runJob(ctrl, w.conf, wrapper, w.logUpdates)
			w.ctrls[jobID] = ctrl
			job.State = ctrl.State()

		case isActive(jobSt) && runSt != Unknown:
			// Job is running, update state.
			job.State = runSt
			if isComplete(runSt) {
				delete(w.ctrls, jobID)
			}

		case jobSt == Canceled:
			// Job is canceled.
			// ctrl.Cancel() is idempotent, so blindly cancel and delete.
			ctrl.Cancel()
			delete(w.ctrls, jobID)

		case jobSt == Unknown && runSt != Unknown:
			// Edge case. There's a ctrl for a non-existant job. Delete it.
			// TODO is it better to leave it? Continue in absence of explicit command principle?
			ctrl.Cancel()
			delete(w.ctrls, jobID)
		case isComplete(jobSt) && isActive(runSt):
			// Edge case. The job is complete but the ctrl is still running.
			// This shouldn't happen but it's better to check for it anyway.
			// TODO better to update job state?
			ctrl.Cancel()
			delete(w.ctrls, jobID)

		default:
			log.Error("Unhandled case during worker reconciliation.",
				"job", job, "ctrl", ctrl)
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
