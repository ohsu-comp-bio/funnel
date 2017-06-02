package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"time"
)

// NewWorker returns a new Worker instance
func NewWorker(conf config.Worker) (*Worker, error) {
	log := logger.Sub("worker", "workerID", conf.ID)
	log.Debug("Worker Config", "config.Worker", conf)

	sched, err := scheduler.NewClient(conf)
	if err != nil {
		return nil, err
	}

	err = util.EnsureDir(conf.WorkDir)
	if err != nil {
		return nil, err
	}

	// Detect available resources at startup
	res := detectResources(conf)
	runners := runSet{}
	timeout := util.NewIdleTimeout(conf.Timeout)
	stop := make(chan struct{})
	state := pbf.WorkerState_UNINITIALIZED
	return &Worker{
		conf, sched, log, res,
		NewDefaultRunner, runners, timeout, stop, state,
	}, nil
}

// Worker is a worker...
type Worker struct {
	conf       config.Worker
	sched      scheduler.Client
	log        logger.Logger
	resources  config.Resources
	newRunner  RunnerFactory
	runners    runSet
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

	// Start task runners. runSet will track task IDs
	// to ensure there's only one runner per ID, so it's ok
	// to call this multiple times with the same task ID.
	for _, id := range r.TaskIds {
		w.runners.Add(id, func(ctx context.Context, id string) {
			r := w.newRunner(w.conf, id)
      r.Run(ctx)
		})
	}

	// Worker data has been updated. Send back to server for database update.
	res := detectResources(w.conf)
	r.Resources = &pbf.Resources{
		Cpus:   res.Cpus,
		RamGb:  res.RamGb,
		DiskGb: res.DiskGb,
	}
	r.State = w.state

	// Merge metadata
	if r.Metadata == nil {
		r.Metadata = map[string]string{}
	}
	for k, v := range w.conf.Metadata {
		r.Metadata[k] = v
	}

	_, err := w.sched.UpdateWorker(context.Background(), r)
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
	w.runners.Stop()
	w.Sync()
	w.sched.Close()
}

// Check if the worker is idle. If so, start the timeout timer.
func (w *Worker) checkIdleTimer() {
	// The worker is idle if there are no task runners.
	idle := w.runners.Count() == 0 && w.state == pbf.WorkerState_ALIVE
	if idle {
		w.timeout.Start()
	} else {
		w.timeout.Stop()
	}
}
