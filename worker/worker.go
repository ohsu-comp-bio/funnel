package worker

import (
	"context"
	"errors"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"time"
)

type logUpdateChan chan *pbf.UpdateExecutorLogsRequest

// NewWorker returns a new Worker instance
func NewWorker(conf config.Worker) (*Worker, error) {
	sched, err := newSchedClient(conf)
	if err != nil {
		return nil, err
	}

	log := logger.New("worker", "workerID", conf.ID)
	log.Debug("Worker Config", "config.Worker", conf)
	res := detectResources(conf.Resources)
	logUpdates := make(logUpdateChan)
	// Tracks active task ctrls: task ID -> TaskControl instance
	ctrls := map[string]TaskControl{}
	timeout := util.NewIdleTimeout(conf.Timeout)
	stop := make(chan struct{})
	state := pbf.WorkerState_UNINITIALIZED
	return &Worker{
		conf, logUpdates, sched, log, res,
		runTask, ctrls, timeout, stop, state,
	}, nil
}

// Worker is a worker...
type Worker struct {
	conf config.Worker
	// Channel for task updates from task runners: stdout/err, ports, etc.
	logUpdates logUpdateChan
	sched      *schedClient
	log        logger.Logger
	resources  config.Resources
	TaskRunner TaskRunner
	Ctrls      map[string]TaskControl
	timeout    util.IdleTimeout
	stop       chan struct{}
	state      pbf.WorkerState
}

// Run runs a worker with the given config. This is responsible for communication
// with the server and starting task runners
func (w *Worker) Run() {
	w.log.Info("Starting worker")
	w.state = pbf.WorkerState_ALIVE
	w.checkConnection()

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

func (w *Worker) checkConnection() {
	_, err := w.sched.GetWorker(context.TODO(), &pbf.GetWorkerRequest{Id: w.conf.ID})

	if err != nil {
		log.Error("Couldn't contact server.", err)
	} else {
		log.Info("Successfully connected to server.")
	}
}

// Sync syncs the worker's state with the server. It reports task state changes,
// handles signals from the server (new task, cancel task, etc), reports resources, etc.
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
	rerr := w.reconcile(r.Tasks)
	if rerr != nil {
		// TODO what's the best behavior here?
		log.Error("Couldn't reconcile worker state.", rerr)
		return
	}

	// Worker data has been updated. Send back to server for database update.
	r.Resources = &pbf.Resources{
		Cpus:   w.resources.Cpus,
		RamGb:  w.resources.RamGb,
		DiskGb: w.resources.DiskGb,
	}
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
	w.state = pbf.WorkerState_GONE
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
	// The worker is idle if there are no task controllers.
	idle := len(w.Ctrls) == 0 && w.state == pbf.WorkerState_ALIVE
	if idle {
		w.timeout.Start()
	} else {
		w.timeout.Stop()
	}
}

func (w *Worker) updateLogs(up *pbf.UpdateExecutorLogsRequest) {
	// UpdateExecutorLogs() is more lightweight than UpdateWorker(),
	// which is why it happens separately and at a different rate.
	err := w.sched.UpdateExecutorLogs(up)
	if err != nil {
		// TODO if the request failed, the task update is lost and the logs
		//      are corrupted. Cache logs to prevent this?
		w.log.Error("Task log update failed", err)
	}
}

// reconcile merges the server state with the worker state:
// - identifies new tasks and starts new runners for them
// - identifies canceled tasks and cancels existing runners
// - updates pbf.Task structs with current task state (running, complete, error, etc)
func (w *Worker) reconcile(tasks map[string]*pbf.TaskWrapper) error {
	// Combine task IDs from response with task IDs from ctrls so we can reconcile
	// both sets below.
	taskIDs := map[string]bool{}
	for taskID := range w.Ctrls {
		taskIDs[taskID] = true
	}
	for taskID := range tasks {
		taskIDs[taskID] = true
	}

	for taskID := range taskIDs {
		taskSt := Unknown
		runSt := Unknown

		ctrl := w.Ctrls[taskID]
		if ctrl != nil {
			runSt = ctrl.State()
		}

		wrapper := tasks[taskID]
		var task *tes.Task
		if wrapper != nil {
			task = wrapper.Task
			taskSt = task.GetState()
		}

		if isComplete(taskSt) {
			delete(w.Ctrls, taskID)
		}

		switch {
		case taskSt == Unknown && runSt == Unknown:
			// Edge case. Shouldn't be here, but log just in case.
			fallthrough
		case isComplete(taskSt) && runSt == Unknown:
			// Edge case. Task is complete and there's no ctrl. Do nothing.
			fallthrough
		case isComplete(taskSt) && runSt == Canceled:
			// Edge case. Task is complete and the ctrl is canceled. Do nothing.
			// This shouldn't happen but it's better to check for it anyway.
			//
			// Log so that these unexpected cases can be explored.
			log.Error("Edge case during worker reconciliation. Recovering.",
				"taskst", taskSt, "runst", runSt)

		case isActive(taskSt) && runSt == Canceled:
			// Edge case. Server says running but ctrl says canceled.
			// Possibly the worker is shutting down due to a local signal
			// and canceled its tasks.
			task.State = runSt

		case taskSt == runSt:
			// States match, do nothing.

		case isActive(taskSt) && runSt == Unknown:
			// Task needs to be started.
			ctrl := NewTaskControl()
			w.TaskRunner(ctrl, w.conf, wrapper, w.logUpdates)
			w.Ctrls[taskID] = ctrl
			task.State = ctrl.State()

		case isActive(taskSt) && runSt != Unknown:
			// Task is running, update state.
			task.State = runSt

		case taskSt == Canceled && runSt != Unknown:
			// Task is canceled.
			// ctrl.Cancel() is idempotent, so blindly cancel and delete.
			ctrl.Cancel()

		case taskSt == Unknown && runSt != Unknown:
			// Edge case. There's a ctrl for a non-existent task. Delete it.
			// TODO is it better to leave it? Continue in absence of explicit command principle?
			ctrl.Cancel()
			delete(w.Ctrls, taskID)

		case isComplete(taskSt) && isActive(runSt):
			// Edge case. The task is complete but the ctrl is still running.
			// This shouldn't happen but it's better to check for it anyway.
			// TODO better to update task state?
			ctrl.Cancel()

		default:
			log.Error("Unhandled case during worker reconciliation.",
				"task", task, "ctrl", ctrl)
			return errors.New("Unhandled case during worker reconciliation")
		}
	}
	return nil
}

func isActive(s tes.State) bool {
	return s == Queued || s == Initializing || s == Running
}

func isComplete(s tes.State) bool {
	return s == Complete || s == Error || s == SystemError
}
