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
	timeout := util.NewIdleTimeout(conf.Timeout)
	state := pbf.WorkerState_UNINITIALIZED
	return &Worker{
		conf, sched, log, res,
		NewDefaultRunner, newRunSet(), timeout, state,
	}, nil
}

// NewNoopWorker returns a new worker that doesn't have any side effects
// (e.g. storage access, docker calls, etc.) which is useful for testing.
func NewNoopWorker(conf config.Worker) (*Worker, error) {
	w, err := NewWorker(conf)
	w.newRunner = NoopRunnerFactory
	return w, err
}

// Worker is a worker...
type Worker struct {
	conf      config.Worker
	sched     scheduler.Client
	log       logger.Logger
	resources config.Resources
	newRunner RunnerFactory
	runners   *runSet
	timeout   util.IdleTimeout
	state     pbf.WorkerState
}

// Run runs a worker with the given config. This is responsible for communication
// with the server and starting task runners
func (w *Worker) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	w.log.Info("Starting worker")
	w.state = pbf.WorkerState_ALIVE
	w.checkConnection(ctx)

	ticker := time.NewTicker(w.conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		case <-w.timeout.Done():
			cancel()
		case <-ctx.Done():
			w.timeout.Stop()

			// The worker gets 10 seconds to do a final sync with the scheduler.
			stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			w.state = pbf.WorkerState_GONE
			w.sync(stopCtx)
			w.sched.Close()

			// The runners get 10 seconds to finish up.
			w.runners.Wait(time.Second * 10)
			return
		case <-ticker.C:
			w.sync(ctx)
			w.checkIdleTimer()
		}
	}
}

func (w *Worker) checkConnection(ctx context.Context) {
	_, err := w.sched.GetWorker(ctx, &pbf.GetWorkerRequest{Id: w.conf.ID})

	if err != nil {
		log.Error("Couldn't contact server.", err)
	} else {
		log.Info("Successfully connected to server.")
	}
}

// sync syncs the worker's state with the server. It reports task state changes,
// handles signals from the server (new task, cancel task, etc), reports resources, etc.
//
// TODO Sync should probably use a channel to sync data access.
//      Probably only a problem for test code, where Sync is called directly.
func (w *Worker) sync(ctx context.Context) {
	r, gerr := w.sched.GetWorker(ctx, &pbf.GetWorkerRequest{Id: w.conf.ID})

	if gerr != nil {
		log.Error("Couldn't get worker state during sync.", gerr)
		return
	}

	// Start task runners. runSet will track task IDs
	// to ensure there's only one runner per ID, so it's ok
	// to call this multiple times with the same task ID.
	for _, id := range r.TaskIds {
		if w.runners.Add(id) {
			go func(id string) {
				r := w.newRunner(w.conf, id)
				r.Run(ctx)
				w.runners.Remove(id)
			}(id)
		}
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

// Check if the worker is idle. If so, start the timeout timer.
func (w *Worker) checkIdleTimer() {
	// The worker is idle if there are no task runners.
	// The worker should not time out if it's not alive (e.g. if it's initializing)
	idle := w.runners.Count() == 0 && w.state == pbf.WorkerState_ALIVE
	if idle {
		w.timeout.Start()
	} else {
		w.timeout.Stop()
	}
}
