package worker

import (
	"context"
	"errors"
	"funnel/config"
	tes "funnel/proto/tes"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/util"
	"time"
)

type logUpdateChan chan *pbf.UpdateJobLogsRequest

// NewWorker returns a new Worker instance
func NewWorker(conf config.Worker) (*Worker, error) {
	sched, err := newSchedClient(conf)
	if err != nil {
		return nil, err
	}

	log := logger.New("worker", "workerID", conf.ID)
	res := detectResources(conf.Resources)
	logUpdates := make(logUpdateChan)
	// Tracks active job ctrls: job ID -> JobControl instance
	ctrls := map[string]JobControl{}
	timeout := util.NewIdleTimeout(conf.Timeout)
	stop := make(chan struct{})
	state := pbf.WorkerState_Uninitialized
	return &Worker{
		conf, logUpdates, sched, log, res,
		runJob, ctrls, timeout, stop, state,
	}, nil
}

// Worker is a worker...
type Worker struct {
	conf config.Worker
	// Channel for job updates from job runners: stdout/err, ports, etc.
	logUpdates logUpdateChan
	sched      *schedClient
	log        logger.Logger
	resources  *pbf.Resources
	JobRunner  JobRunner
	Ctrls      map[string]JobControl
	timeout    util.IdleTimeout
	stop       chan struct{}
	state      pbf.WorkerState
}

// Run runs a worker with the given config. This is responsible for communication
// with the server and starting job runners
func (w *Worker) Run() {
	w.log.Info("Starting worker")
	w.state = pbf.WorkerState_Alive

	ticker := time.NewTicker(w.conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case up := <-w.logUpdates:
			w.updateLogs(up)
		case <-ticker.C:
			w.Sync()
			w.checkIdleTimer()
		case <-w.timeout.Done():
			// Worker timeout reached. Shutdown.
			w.Stop()
			return
		}
	}
}

// Sync syncs the worker's state with the server. It reports job state changes,
// handles signals from the server (new job, cancel job, etc), reports resources, etc.
//
// TODO Sync should probably use a channel to sync data access.
//      Probably only a problem for test code, where Sync is called directly.
func (w *Worker) Sync() {
	r, gerr := w.sched.GetWorker(context.TODO(), &pbf.GetWorkerRequest{Id: w.conf.ID})

	if gerr != nil {
		log.Error("Couldn't get worker state during sync.", gerr)
		return
	}

	// Reconcile server state with worker state.
	rerr := w.reconcile(r.Jobs)
	if rerr != nil {
		// TODO what's the best behavior here?
		log.Error("Couldn't reconcile worker state.", rerr)
		return
	}

	// Worker data has been updated. Send back to server for database update.
	r.Resources = w.resources
	r.State = w.state

	// Merge metadata
	if r.Metadata == nil {
		r.Metadata = map[string]string{}
	}
	for k, v := range w.conf.Metadata {
		r.Metadata[k] = v
	}

	_, err := w.sched.UpdateWorker(r)
	if err != nil {
		log.Error("Couldn't save worker update. Recovering.", err)
	}
}

// Stop stops the worker
// TODO need a way to shut the worker down from the server/scheduler.
func (w *Worker) Stop() {
	w.state = pbf.WorkerState_Gone
	close(w.stop)
	w.timeout.Stop()
	for _, ctrl := range w.Ctrls {
		ctrl.Cancel()
	}
	w.Sync()
	w.sched.Close()
}

// Check if the worker is idle. If so, start the timeout timer.
func (w *Worker) checkIdleTimer() {
	// The worker is idle if there are no job controllers.
	idle := len(w.Ctrls) == 0 && w.state == pbf.WorkerState_Alive
	if idle {
		w.timeout.Start()
	} else {
		w.timeout.Stop()
	}
}

func (w *Worker) updateLogs(up *pbf.UpdateJobLogsRequest) {
	// UpdateJobLogs() is more lightweight than UpdateWorker(),
	// which is why it happens separately and at a different rate.
	err := w.sched.UpdateJobLogs(up)
	if err != nil {
		// TODO if the request failed, the job update is lost and the logs
		//      are corrupted. Cache logs to prevent this?
		w.log.Error("Job log update failed", err)
	}
}

// reconcile merges the server state with the worker state:
// - identifies new jobs and starts new runners for them
// - identifies canceled jobs and cancels existing runners
// - updates pbf.Job structs with current job state (running, complete, error, etc)
func (w *Worker) reconcile(jobs map[string]*pbf.JobWrapper) error {
	var (
		Unknown  = tes.State_Unknown
		Canceled = tes.State_Canceled
	)

	// Combine job IDs from response with job IDs from ctrls so we can reconcile
	// both sets below.
	jobIDs := map[string]bool{}
	for jobID := range w.Ctrls {
		jobIDs[jobID] = true
	}
	for jobID := range jobs {
		jobIDs[jobID] = true
	}

	for jobID := range jobIDs {
		jobSt := tes.State_Unknown
		runSt := tes.State_Unknown

		ctrl := w.Ctrls[jobID]
		if ctrl != nil {
			runSt = ctrl.State()
		}

		wrapper := jobs[jobID]
		var job *tes.Job
		if wrapper != nil {
			job = wrapper.Job
			jobSt = job.GetState()
		}

		if isComplete(jobSt) {
			delete(w.Ctrls, jobID)
		}

		switch {
		case jobSt == Unknown && runSt == Unknown:
			// Edge case. Shouldn't be here, but log just in case.
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

		case isActive(jobSt) && runSt == Canceled:
			// Edge case. Server says running but ctrl says canceled.
			// Possibly the worker is shutting down due to a local signal
			// and canceled its jobs.
			job.State = runSt

		case jobSt == runSt:
			// States match, do nothing.

		case isActive(jobSt) && runSt == Unknown:
			// Job needs to be started.
			ctrl := NewJobControl()
			w.JobRunner(ctrl, w.conf, wrapper, w.logUpdates)
			w.Ctrls[jobID] = ctrl
			job.State = ctrl.State()

		case isActive(jobSt) && runSt != Unknown:
			// Job is running, update state.
			job.State = runSt

		case jobSt == Canceled && runSt != Unknown:
			// Job is canceled.
			// ctrl.Cancel() is idempotent, so blindly cancel and delete.
			ctrl.Cancel()

		case jobSt == Unknown && runSt != Unknown:
			// Edge case. There's a ctrl for a non-existent job. Delete it.
			// TODO is it better to leave it? Continue in absence of explicit command principle?
			ctrl.Cancel()
			delete(w.Ctrls, jobID)

		case isComplete(jobSt) && isActive(runSt):
			// Edge case. The job is complete but the ctrl is still running.
			// This shouldn't happen but it's better to check for it anyway.
			// TODO better to update job state?
			ctrl.Cancel()

		default:
			log.Error("Unhandled case during worker reconciliation.",
				"job", job, "ctrl", ctrl)
			return errors.New("Unhandled case during worker reconciliation")
		}
	}
	return nil
}

func isActive(s tes.State) bool {
	return s == tes.State_Queued || s == tes.State_Initializing || s == tes.State_Running
}

func isComplete(s tes.State) bool {
	return s == tes.State_Complete || s == tes.State_Error || s == tes.State_SystemError
}
