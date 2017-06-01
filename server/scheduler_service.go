package server

// TODO put the boltdb implementation in a separate package
//      so that users can import pluggable backends

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"time"
)

// State variables for convenience
const (
	Unknown      = tes.State_UNKNOWN
	Queued       = tes.State_QUEUED
	Running      = tes.State_RUNNING
	Paused       = tes.State_PAUSED
	Complete     = tes.State_COMPLETE
	Error        = tes.State_ERROR
	SystemError  = tes.State_SYSTEM_ERROR
	Canceled     = tes.State_CANCELED
	Initializing = tes.State_INITIALIZING
)

// UpdateWorker is an RPC endpoint that is used by workers to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the worker, such as canceled tasks.
func (taskBolt *TaskBolt) UpdateWorker(ctx context.Context, req *pbf.Worker) (*pbf.UpdateWorkerResponse, error) {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		return updateWorker(tx, req)
	})
	resp := &pbf.UpdateWorkerResponse{}
	return resp, err
}

func updateWorker(tx *bolt.Tx, req *pbf.Worker) error {
	// Get worker
	worker := getWorker(tx, req.Id)

	if worker.Version != 0 && req.Version != 0 && worker.Version != req.Version {
		return errors.New("Version outdated")
	}

	worker.LastPing = time.Now().Unix()
	worker.State = req.GetState()

	if req.Resources != nil {
		if worker.Resources == nil {
			worker.Resources = &pbf.Resources{}
		}
		// Merge resources
		if req.Resources.Cpus > 0 {
			worker.Resources.Cpus = req.Resources.Cpus
		}
		if req.Resources.RamGb > 0 {
			worker.Resources.RamGb = req.Resources.RamGb
		}
		if req.Resources.DiskGb > 0 {
			worker.Resources.DiskGb = req.Resources.DiskGb
		}
	}

	// Reconcile worker's task states with database
	for _, wrapper := range req.Tasks {
		// TODO test transition to self a noop
		task := wrapper.Task
		err := transitionTaskState(tx, task.Id, task.State)
		// TODO what's the proper behavior of an error?
		//      this is just ignoring the error, but it will happen again
		//      on the next update.
		//      need to resolve the conflicting states.
		//      Additionally, returning an error here will fail the db transaction,
		//      preventing all updates to this worker for all tasks.
		if err != nil {
			return err
		}

		// If the worker has acknowledged that the task is complete,
		// unlink the task from the worker.
		switch task.State {
		case Canceled, Complete, Error, SystemError:
			key := append([]byte(req.Id), []byte(task.Id)...)
			tx.Bucket(WorkerTasks).Delete(key)
		}
	}

	for k, v := range req.Metadata {
		worker.Metadata[k] = v
	}

	// TODO move to on-demand helper. i.e. don't store in DB
	updateAvailableResources(tx, worker)
	worker.Version = time.Now().Unix()
	putWorker(tx, worker)
	return nil
}

// AssignTask assigns a task to a worker. This updates the task state to Initializing,
// and updates the worker (calls UpdateWorker()).
func (taskBolt *TaskBolt) AssignTask(t *tes.Task, w *pbf.Worker) {
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		// TODO this is important! write a test for this line.
		//      when a task is assigned, its state is immediately Initializing
		//      even before the worker has received it.
		transitionTaskState(tx, t.Id, tes.State_INITIALIZING)
		taskIDBytes := []byte(t.Id)
		workerIDBytes := []byte(w.Id)
		// TODO the database needs tests for this stuff. Getting errors during dev
		//      because it's easy to forget to link everything.
		key := append(workerIDBytes, taskIDBytes...)
		tx.Bucket(WorkerTasks).Put(key, taskIDBytes)
		tx.Bucket(TaskWorker).Put(taskIDBytes, workerIDBytes)

		err := updateWorker(tx, w)
		if err != nil {
			return err
		}
		return nil
	})
}

// TODO include active ports. maybe move Available out of the protobuf message
//      and expect this helper to be used?
func updateAvailableResources(tx *bolt.Tx, worker *pbf.Worker) {
	// Calculate available resources
	a := pbf.Resources{
		Cpus:   worker.GetResources().GetCpus(),
		RamGb:  worker.GetResources().GetRamGb(),
		DiskGb: worker.GetResources().GetDiskGb(),
	}
	for taskID := range worker.Tasks {
		t := getTask(tx, taskID)
		res := t.GetResources()

		// Cpus are represented by an unsigned int, and if we blindly
		// subtract it will rollover to a very large number. So check first.
		rcpus := res.GetCpuCores()
		if rcpus >= a.Cpus {
			a.Cpus = 0
		} else {
			a.Cpus -= rcpus
		}

		a.RamGb -= res.GetRamGb()

		if a.Cpus < 0 {
			a.Cpus = 0
		}
		if a.RamGb < 0.0 {
			a.RamGb = 0.0
		}
	}
	worker.Available = &a
}

// GetWorker gets a worker
func (taskBolt *TaskBolt) GetWorker(ctx context.Context, req *pbf.GetWorkerRequest) (*pbf.Worker, error) {
	var worker *pbf.Worker
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		worker = getWorker(tx, req.Id)
		return nil
	})
	return worker, err
}

// CheckWorkers is used by the scheduler to check for dead/gone workers.
// This is not an RPC endpoint
func (taskBolt *TaskBolt) CheckWorkers() error {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(Workers)
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			worker := &pbf.Worker{}
			proto.Unmarshal(v, worker)

			if worker.State == pbf.WorkerState_GONE {
				tx.Bucket(Workers).Delete(k)
				continue
			}

			if worker.LastPing == 0 {
				// This shouldn't be happening, because workers should be
				// created with LastPing, but give it the benefit of the doubt
				// and leave it alone.
				continue
			}

			lastPing := time.Unix(worker.LastPing, 0)
			d := time.Since(lastPing)

			if worker.State == pbf.WorkerState_UNINITIALIZED ||
				worker.State == pbf.WorkerState_INITIALIZING {

				// The worker is initializing, which has a more liberal timeout.
				if d > taskBolt.conf.WorkerInitTimeout {
					// Looks like the worker failed to initialize. Mark it dead
					worker.State = pbf.WorkerState_DEAD
				}
			} else if d > taskBolt.conf.WorkerPingTimeout {
				// The worker is stale/dead
				worker.State = pbf.WorkerState_DEAD
			} else {
				worker.State = pbf.WorkerState_ALIVE
			}
			// TODO when to delete workers from the database?
			//      is dead worker deletion an automatic garbage collection process?
			putWorker(tx, worker)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ListWorkers is an API endpoint that returns a list of workers.
func (taskBolt *TaskBolt) ListWorkers(ctx context.Context, req *pbf.ListWorkersRequest) (*pbf.ListWorkersResponse, error) {
	resp := &pbf.ListWorkersResponse{}
	resp.Workers = []*pbf.Worker{}

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(Workers)
		c := bucket.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			worker := getWorker(tx, string(k))
			resp.Workers = append(resp.Workers, worker)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Look for an auth token related to the given task ID.
func getTaskAuth(tx *bolt.Tx, taskID string) string {
	idBytes := []byte(taskID)
	var auth string
	data := tx.Bucket(TaskAuthBucket).Get(idBytes)
	if data != nil {
		auth = string(data)
	}
	return auth
}

func transitionTaskState(tx *bolt.Tx, id string, state tes.State) error {
	idBytes := []byte(id)
	current := getTaskState(tx, id)

	switch current {
	case state:
		// Current state matches target state. Do nothing.
		return nil

	case Complete, Error, SystemError, Canceled:
		// Current state is a terminal state, can't do that.
		err := errors.New("Invalid state change")
		log.Error("Cannot change state of a task already in a terminal state",
			"error", err,
			"current", current,
			"requested", state)
		return err
	}

	switch state {
	case Canceled, Complete, Error, SystemError:
		// Remove from queue
		tx.Bucket(TasksQueued).Delete(idBytes)

	case Running, Initializing:
		if current != Unknown && current != Queued && current != Initializing {
			log.Error("Unexpected transition", "current", current, "requested", state)
			return errors.New("Unexpected transition to Initializing")
		}
		tx.Bucket(TasksQueued).Delete(idBytes)

	case Unknown, Paused:
		log.Error("Unimplemented task state", "state", state)
		return errors.New("Unimplemented task state")

	case Queued:
		log.Error("Can't transition to Queued state")
		return errors.New("Can't transition to Queued state")
	default:
		log.Error("Unknown task state", "state", state)
		return errors.New("Unknown task state")
	}

	tx.Bucket(TaskState).Put(idBytes, []byte(state.String()))
	log.Info("Set task state", "taskID", id, "state", state.String())
	return nil
}

// UpdateTaskLogs is an internal API endpoint that allows the worker to update
// task logs (start time, end time, output files, etc).
func (taskBolt *TaskBolt) UpdateTaskLogs(ctx context.Context, req *pbf.UpdateTaskLogsRequest) (*pbf.UpdateTaskLogsResponse, error) {
	log.Debug("Update task logs", req)

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		tasklog := &tes.TaskLog{}

		// Try to load existing task log
		b := tx.Bucket(TasksLog).Get([]byte(req.Id))
		if b != nil {
			proto.Unmarshal(b, tasklog)
		}

		if req.TaskLog.StartTime != "" {
			tasklog.StartTime = req.TaskLog.StartTime
		}

		if req.TaskLog.EndTime != "" {
			tasklog.EndTime = req.TaskLog.EndTime
		}

		if req.TaskLog.Outputs != nil {
			tasklog.Outputs = req.TaskLog.Outputs
		}

		if req.TaskLog.Metadata != nil {
			if tasklog.Metadata == nil {
				tasklog.Metadata = map[string]string{}
			}
			for k, v := range req.TaskLog.Metadata {
				tasklog.Metadata[k] = v
			}
		}

		logbytes, _ := proto.Marshal(tasklog)
		tx.Bucket(TasksLog).Put([]byte(req.Id), logbytes)
		return nil
	})
	return &pbf.UpdateTaskLogsResponse{}, err
}

// UpdateExecutorLogs is an API endpoint that updates the logs of a task.
// This is used by workers to communicate task updates to the server.
func (taskBolt *TaskBolt) UpdateExecutorLogs(ctx context.Context, req *pbf.UpdateExecutorLogsRequest) (*pbf.UpdateExecutorLogsResponse, error) {
	log.Debug("Update task executor logs", req)

	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bL := tx.Bucket(ExecutorLogs)

		// max size (bytes) for stderr and stdout streams to keep in db
		max := taskBolt.conf.MaxExecutorLogSize
		key := []byte(fmt.Sprint(req.Id, req.Step))

		if req.Log != nil {
			// Check if there is an existing task log
			o := bL.Get(key)
			if o != nil {
				log.Debug("UPDTA")
				// There is an existing log in the DB, load it
				existing := &tes.ExecutorLog{}
				// max bytes to be stored in the db
				proto.Unmarshal(o, existing)

				stdout := []byte(existing.Stdout + req.Log.Stdout)
				stderr := []byte(existing.Stderr + req.Log.Stderr)

				// Trim the stdout/err logs to the max size if needed
				if len(stdout) > max {
					stdout = stdout[:max]
				}
				if len(stderr) > max {
					stderr = stderr[:max]
				}

				req.Log.Stdout = string(stdout)
				req.Log.Stderr = string(stderr)

				// Merge the updates into the existing.
				proto.Merge(existing, req.Log)
				// existing is updated, so set that to req.Log which will get saved below.
				req.Log = existing
			}

			// Save the updated log
			logbytes, _ := proto.Marshal(req.Log)
			tx.Bucket(ExecutorLogs).Put(key, logbytes)
		}

		return nil
	})
	return &pbf.UpdateExecutorLogsResponse{}, nil
}
