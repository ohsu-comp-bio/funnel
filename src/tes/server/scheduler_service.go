package server

// TODO put the boltdb implementation in a separate package
//      so that users can import pluggable backends

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
	"time"
)

func (taskBolt *TaskBolt) UpdateWorker(ctx context.Context, req *pbr.UpdateWorkerRequest) (*pbr.UpdateWorkerResponse, error) {

	resp := &pbr.UpdateWorkerResponse{}

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		// Get worker
		worker, werr := getWorker(tx, req.Id)
		if werr != nil {
			return werr
		}

		// Update worker metadata
		worker.LastPing = time.Now().Unix()
		worker.Resources = req.Resources

		// Reconcile worker's job states with database
		for jobID, state := range req.States {
			switch state {

			case pbe.State_Initializing, pbe.State_Running:
				// The worker has acknowledged receiving the job, so remove it from Assigned.
				delete(worker.Assigned, jobID)
				worker.Active[jobID] = true
				// If the job was canceled, add that signal to the response.
				//
				// Don't remove canceled jobs from the worker right away. Wait until
				// the next update, which acknowledges the worker has received the cancel.
				canceled := getJobState(tx, jobID) == pbe.State_Canceled
				if canceled {
					resp.Canceled = append(resp.Canceled, jobID)
				}

			// Terminal states. Update state in database.
			case pbe.State_Error, pbe.State_Complete, pbe.State_Canceled:
				delete(worker.Assigned, jobID)
				delete(worker.Active, jobID)

			default:
				log.Error("Unknown job state during worker update", "state", state)
				continue
			}

			err := transitionJobState(tx, jobID, state)
			// TODO what's the proper behavior of an error?
			//      this is just ignoring the error, but it will happen again on the next update.
			//      need to resolve the conflicting states.
			//      Additionally, returning an error here will fail the db transaction,
			//      preventing all updates to this worker for all jobs.
			if err != nil {
				return err
			}
		}

		// New jobs have been assigned, add these to the update response.
		//
		// Don't remove these from worker.Assigned right away. It's possible that
		// the worker won't receive the update response. Wait until the worker sends
		// and update including the jobID.
		for jobID := range worker.Assigned {
			resp.Assigned = append(resp.Assigned, &pbr.Assignment{
				Job:  getJob(tx, jobID),
				Auth: getJobAuth(tx, jobID),
			})
		}

		// Save the worker
		putWorker(tx, worker)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (taskBolt *TaskBolt) GetWorkers(ctx context.Context, req *pbr.GetWorkersRequest) (*pbr.GetWorkersResponse, error) {
	resp := &pbr.GetWorkersResponse{}
	resp.Workers = []*pbr.Worker{}

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(Workers)
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			worker := &pbr.Worker{}
			proto.Unmarshal(v, worker)
			resp.Workers = append(resp.Workers, worker)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Look for an auth token related to the given job ID.
func getJobAuth(tx *bolt.Tx, jobID string) string {
	idBytes := []byte(jobID)
	var auth string
	data := tx.Bucket(TaskAuthBucket).Get(idBytes)
	if data != nil {
		auth = string(data)
	}
	return auth
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
		worker.Assigned[id] = true
		putWorker(tx, worker)

		// TODO do we want an "Assigned" state?
		err := transitionJobState(tx, id, pbe.State_Initializing)
		if err != nil {
			return err
		}
		return nil
	})
}

func transitionJobState(tx *bolt.Tx, id string, state pbe.State) error {
	idBytes := []byte(id)

	var (
		Unknown      = pbe.State_Unknown
		Queued       = pbe.State_Queued
		Running      = pbe.State_Running
		Paused       = pbe.State_Paused
		Complete     = pbe.State_Complete
		Error        = pbe.State_Error
		SystemError  = pbe.State_SystemError
		Canceled     = pbe.State_Canceled
		Initializing = pbe.State_Initializing
	)

	current := getJobState(tx, id)

	switch current {
	// Current state matches target state. Do nothing.
	case state:
		return nil

	// Current state is a terminal state, can't do that.
	case Complete, Error, SystemError, Canceled:
		err := errors.New("Invalid state change")
		log.Error("Cannot change state of a job already in a terminal state",
			"error", err,
			"current", current,
			"requested", state)
		return err
	}

	switch state {
	case Canceled, Complete, Error, SystemError:
		idBytes := []byte(id)
		// Remove from queue
		tx.Bucket(JobsQueued).Delete(idBytes)

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

// UpdateJobStatus updates the status of a job, including state and logs.
// This is an RPC endpoint.
// This is used by workers to communicate job updates to the server.
func (taskBolt *TaskBolt) UpdateJobLogs(ctx context.Context, req *pbr.UpdateJobLogsRequest) (*pbr.UpdateJobLogsResponse, error) {

	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bL := tx.Bucket(JobsLog)

		// max size (bytes) for stderr and stdout streams to keep in db
		max := taskBolt.conf.MaxJobLogSize
		key := []byte(fmt.Sprint(req.Id, req.Step))

		if req.Log != nil {
			// Check if there is an existing job log
			o := bL.Get(key)
			if o != nil {
				// There is an existing log in the DB, load it
				existing := &pbe.JobLog{}
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
			tx.Bucket(JobsLog).Put(key, logbytes)
		}

		return nil
	})
	return &pbr.UpdateJobLogsResponse{}, nil
}

// GetQueueInfo returns a stream of queue info
// This is an RPC endpoint.
// TODO why doesn't this take Context as the first argument?
// TODO I don't think this is actually used.
func (taskBolt *TaskBolt) GetQueueInfo(request *pbr.QueuedTaskInfoRequest, server pbr.Scheduler_GetQueueInfoServer) error {
	ch := make(chan *pbe.Task)
	log.Debug("GetQueueInfo called")

	// TODO handle DB errors
	go taskBolt.db.View(func(tx *bolt.Tx) error {
		defer close(ch)

		bt := tx.Bucket(TaskBucket)
		bq := tx.Bucket(JobsQueued)
		c := bq.Cursor()
		var count int32
		for k, v := c.First(); k != nil && count < request.MaxTasks; k, v = c.Next() {
			if string(v) == pbe.State_Queued.String() {
				v := bt.Get(k)
				out := pbe.Task{}
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
		server.Send(&pbr.QueuedTaskInfo{Inputs: inputs, Resources: m.Resources})
	}

	return nil
}
