package server

// TODO put the boltdb implementation in a separate package
//      so that users can import pluggable backends

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"tes/config"
	"tes/ga4gh"
	"tes/server/proto"
)

// GetJobToRun returns a queued job for a worker to run.
// This is an RPC endpoint.
// This is used by workers to request work.
func (taskBolt *TaskBolt) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResponse, error) {
  log.Debug("GetJobToRun called", "workerID", request.Worker.Id)

  var task *ga4gh_task_exec.Task
	var jobID, authToken string

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		worker, werr := getWorker(tx, request.Worker.Id)
    if werr != nil {
      return werr
    }

		if len(worker.QueuedJobs) == 0 {
      log.Debug("Worker has empty job queue", "worker", worker)
			return nil
		}

		// Shift job from queued to active
		jobID, worker.QueuedJobs = worker.QueuedJobs[0], worker.QueuedJobs[1:]
		worker.ActiveJobs = append(worker.ActiveJobs, jobID)
		putWorker(tx, worker)

		bOp := tx.Bucket(TaskBucket)
		authBkt := tx.Bucket(TaskAuthBucket)

		// Get the task
    task = &ga4gh_task_exec.Task{}
		v := bOp.Get([]byte(jobID))
		proto.Unmarshal(v, task)

		// Look for an auth token related to this task
		tok := authBkt.Get([]byte(jobID))
		if tok != nil {
			authToken = string(tok)
		}
		return nil
	})

  if err != nil {
    return nil, err
  }

	// No task was found. Respond accordingly.
	if task == nil {
		return &ga4gh_task_ref.JobResponse{}, nil
	}

	job := &ga4gh_task_exec.Job{
		JobID: jobID,
		Task:  task,
	}

	return &ga4gh_task_ref.JobResponse{Job: job, Auth: authToken}, nil
}

func getWorker(tx *bolt.Tx, id string) (*Worker, error) {
  pb := &ga4gh_task_ref.Worker{
    Id: id,
  }
  worker := &Worker{pb}

	data := tx.Bucket(Workers).Get([]byte(id))
	if data != nil {
	  proto.Unmarshal(data, pb)
	}
	return worker, nil
}

func putWorker(tx *bolt.Tx, worker *Worker) error {
  log.Debug("Put worker", "worker", worker)
	bw := tx.Bucket(Workers)
	data, _ := proto.Marshal(worker.Worker)
	bw.Put([]byte(worker.Id), data)
	return nil
}

// AssignJob assigns a job to a worker.
// This is NOT an RPC endpoint.
func (taskBolt *TaskBolt) AssignJob(id string, workerID string) error {
  return taskBolt.db.Update(func(tx *bolt.Tx) error {
		// Append job id to worker's queued jobs
		worker, werr := getWorker(tx, workerID)
    if werr != nil {
      return werr
    }
		worker.QueuedJobs = append(worker.QueuedJobs, id)
		putWorker(tx, worker)

		err := transitionJobState(tx, id, ga4gh_task_exec.State_Running)
		if err != nil {
			return err
		}
		// Link job to worker
		tx.Bucket(JobWorker).Put([]byte(id), []byte(workerID))
		return nil
	})
	return nil
}

func transitionJobState(tx *bolt.Tx, id string, state ga4gh_task_exec.State) error {
	idBytes := []byte(id)

	var (
		Unknown      = ga4gh_task_exec.State_Unknown
		Queued       = ga4gh_task_exec.State_Queued
		Running      = ga4gh_task_exec.State_Running
		Paused       = ga4gh_task_exec.State_Paused
		Complete     = ga4gh_task_exec.State_Complete
		Error        = ga4gh_task_exec.State_Error
		SystemError  = ga4gh_task_exec.State_SystemError
		Canceled     = ga4gh_task_exec.State_Canceled
		Initializing = ga4gh_task_exec.State_Initializing
	)

	current := getJobState(tx, id)

	if current == Complete || current == Error || current == SystemError || current == Canceled {
		err := errors.New("Invalid state change")
		log.Error("Cannot change state of a job already in a terminal state",
			"error", err,
			"current", current,
			"requested", state)
		return err
	}

	if current == state {
		return nil
	}

	switch state {
	case Canceled, Complete, Error, SystemError:
		clearJob(tx, id)

	case Running, Initializing:
		if current != Unknown && current != Queued && current != Initializing {
			log.Error("Unexpected transition", "current", current, "requested", state)
			return errors.New("Unexpected transition to Initializing")
		}
		tx.Bucket(JobsQueued).Delete(idBytes)

	case Unknown, Paused:
		log.Error("Unimplemented job state", "state", state)
		return errors.New("Unimplemented job state")

	case Queued:
		log.Error("Can't transition to Queued state")
		return errors.New("Can't transition to Queued state")
	default:
		log.Error("Unknown job state", "state", state)
		return errors.New("Unknown job state")
	}

	tx.Bucket(JobState).Put(idBytes, []byte(state.String()))
	log.Info("Set job state", "jobID", id, "state", state.String())
	return nil
}

// clearJob helps remove a job from the various job state tracking buckets,
// e.g. JobsQueued, JobsWorkers, Worker.ActiveJobs, etc.
// Use this when you need to put a job into a terminal state and need to clean
// up it's state in all these buckets.
func clearJob(tx *bolt.Tx, id string) {
	idBytes := []byte(id)
	// Remove from queue
	tx.Bucket(JobsQueued).Delete(idBytes)
	// Remove from job ID -> worker mapping
	workerID := tx.Bucket(JobWorker).Get(idBytes)
	tx.Bucket(JobWorker).Delete(idBytes)
	// Remove from worker
	worker, _ := getWorker(tx, string(workerID))
	worker.RemoveJob(id)
	putWorker(tx, worker)
}

// UpdateJobStatus updates the status of a job, including state and logs.
// This is an RPC endpoint.
// This is used by workers to communicate job updates to the server.
func (taskBolt *TaskBolt) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobID, error) {
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bL := tx.Bucket(JobsLog)

		// max size (bytes) for stderr and stdout streams to keep in db
		// TODO make configurable
		max := 100000
		if stat.Log != nil {
			out := &ga4gh_task_exec.JobLog{}
			o := bL.Get([]byte(fmt.Sprint(stat.Id, stat.Step)))
			if o != nil {
				var jlog ga4gh_task_exec.JobLog
				// max bytes to be stored in the db
				proto.Unmarshal(o, &jlog)
				out = &jlog
				stdout := []byte(out.Stdout + stat.Log.Stdout)
				stderr := []byte(out.Stderr + stat.Log.Stderr)
				if len(stdout) > max {
					stdout = stdout[len(stdout)-max:]
				}
				if len(stderr) > max {
					stderr = stderr[len(stderr)-max:]
				}
				out.Stdout = string(stdout)
				out.Stderr = string(stderr)
			} else {
				out = stat.Log
			}
			dL, _ := proto.Marshal(out)
			bL.Put([]byte(fmt.Sprint(stat.Id, stat.Step)), dL)
		}

		return nil
	})
	return &ga4gh_task_exec.JobID{Value: stat.Id}, nil
}

type Worker struct {
	*ga4gh_task_ref.Worker
}

func (w *Worker) RemoveJob(id string) {
	// Remove job from w job lists
	for i, jobID := range w.ActiveJobs {
		if jobID == id {
			w.ActiveJobs = append(w.ActiveJobs[:i], w.ActiveJobs[i+1:]...)
			break
		}
	}
	// Remove job from w job lists
	for i, jobID := range w.QueuedJobs {
		if jobID == id {
			w.QueuedJobs = append(w.QueuedJobs[:i], w.QueuedJobs[i+1:]...)
			break
		}
	}
}

// JobComplete is used by the worker to notify the scheduler that the job is complete.
func (taskBolt *TaskBolt) JobComplete(ctx context.Context, req *ga4gh_task_ref.JobCompleteRequest) (*ga4gh_task_ref.JobCompleteResponse, error) {
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		if req.Failed {
			transitionJobState(tx, req.Id, ga4gh_task_exec.State_Error)
		} else {
			transitionJobState(tx, req.Id, ga4gh_task_exec.State_Complete)
		}

		// Look up worker ID
		workerID := tx.Bucket(JobWorker).Get([]byte(req.Id))
		// Remove job from worker
		worker, werr := getWorker(tx, string(workerID))
    if werr != nil {
      return werr
    }
		worker.RemoveJob(req.Id)
		putWorker(tx, worker)
		return nil
	})
	return &ga4gh_task_ref.JobCompleteResponse{}, nil
}

// WorkerPing tells the server that a worker is alive.
// This is an RPC endpoint.
// This is currently unimplemented. TODO
func (taskBolt *TaskBolt) WorkerPing(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.WorkerInfo, error) {
	log.Debug("Worker ping")
	return info, nil
}

// GetQueueInfo returns a stream of queue info
// This is an RPC endpoint.
// TODO why doesn't this take Context as the first argument?
// TODO I don't think this is actually used.
func (taskBolt *TaskBolt) GetQueueInfo(request *ga4gh_task_ref.QueuedTaskInfoRequest, server ga4gh_task_ref.Scheduler_GetQueueInfoServer) error {
	ch := make(chan *ga4gh_task_exec.Task)
	log.Debug("GetQueueInfo called")

	// TODO handle DB errors
	go taskBolt.db.View(func(tx *bolt.Tx) error {
		defer close(ch)

		bt := tx.Bucket(TaskBucket)
		bq := tx.Bucket(JobsQueued)
		c := bq.Cursor()
		var count int32
		for k, v := c.First(); k != nil && count < request.MaxTasks; k, v = c.Next() {
			if string(v) == ga4gh_task_exec.State_Queued.String() {
				v := bt.Get(k)
				out := ga4gh_task_exec.Task{}
				proto.Unmarshal(v, &out)
				ch <- &out
			}
		}

		return nil
	})

	for m := range ch {
		inputs := make([]string, 0, len(m.Inputs))
		for _, i := range m.Inputs {
			inputs = append(inputs, i.Location)
		}
		server.Send(&ga4gh_task_ref.QueuedTaskInfo{Inputs: inputs, Resources: m.Resources})
	}

	return nil
}

// GetServerConfig returns information about the server configuration.
// This is an RPC endpoint.
func (taskBolt *TaskBolt) GetServerConfig(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*config.Config, error) {
	return &taskBolt.serverConfig, nil
}

// GetJobState returns the state of a job, given a job ID.
// This is an RPC endpoint.
func (taskBolt *TaskBolt) GetJobState(ctx context.Context, id *ga4gh_task_exec.JobID) (*ga4gh_task_exec.JobDesc, error) {
	log.Debug("GetJobState called")
	var state ga4gh_task_exec.State
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		state = getJobState(tx, id.Value)
		return nil
	})
	jobDesc := &ga4gh_task_exec.JobDesc{
		JobID: id.Value,
		State: state,
	}
	return jobDesc, err
}
