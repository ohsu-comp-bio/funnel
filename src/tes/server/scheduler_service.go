package server

// TODO put the boltdb implementation in a separate package
//      so that users can import pluggable backends

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"tes"
	"tes/ga4gh"
	"tes/server/proto"
)

// GetJobToRun returns a queued job for a worker to run.
// This is an RPC endpoint.
// This is used by workers to request work.
func (taskBolt *TaskBolt) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResponse, error) {
	var task *ga4gh_task_exec.Task
	var jobID string
	authToken := ""

	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bOp := tx.Bucket(TaskBucket)
		bw := tx.Bucket(WorkerJobs)
		authBkt := tx.Bucket(TaskAuthBucket)

		k := bw.Get([]byte(request.Worker.Id))
		if k != nil {
			// Get the task
			v := bOp.Get(k)
			task = &ga4gh_task_exec.Task{}
			jobID = string(k)
			proto.Unmarshal(v, task)
			// Update the job state to "Running"

			// Look for an auth token related to this task
			tok := authBkt.Get([]byte(k))
			if tok != nil {
				authToken = string(tok)
			}
		}
		return nil
	})
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

// AssignJob assigns a job to a worker.
// This is NOT an RPC endpoint.
func (taskBolt *TaskBolt) AssignJob(id string, workerID string) error {
	running := []byte(ga4gh_task_exec.State_Running.String())
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JobsActive)
		bc := tx.Bucket(JobsComplete)
		bq := tx.Bucket(JobsQueued)
		bw := tx.Bucket(WorkerJobs)
		bjw := tx.Bucket(JobWorker)
		k := []byte(id)
		w := []byte(workerID)
		bc.Delete(k)
		bq.Delete(k)
		ba.Put(k, running)
		bjw.Put(k, w)
		bw.Put(w, k)
		return nil
	})
	return nil
}

// UpdateJobStatus updates the status of a job, including state and logs.
// This is an RPC endpoint.
// This is used by workers to communicate job updates to the server.
func (taskBolt *TaskBolt) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobID, error) {
	log := log.WithFields("jobID", stat.Id)

	taskBolt.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JobsActive)
		bc := tx.Bucket(JobsComplete)
		bL := tx.Bucket(JobsLog)
		bw := tx.Bucket(WorkerJobs)
		bjw := tx.Bucket(JobWorker)

		// max size (bytes) for stderr and stdout streams to keep in db
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

		switch stat.State {
		case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error:
			log.Debug("Job state change", "state", stat.State)
			workerID := bjw.Get([]byte(stat.Id))
			bjw.Delete([]byte(stat.Id))
			ba.Delete([]byte(stat.Id))
			bw.Delete([]byte(workerID))
			bc.Put([]byte(stat.Id), []byte(stat.State.String()))
		case ga4gh_task_exec.State_Initializing, ga4gh_task_exec.State_Running:
			log.Debug("Job state", "state", stat.State)
		}
		return nil
	})
	return &ga4gh_task_exec.JobID{Value: stat.Id}, nil
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
func (taskBolt *TaskBolt) GetServerConfig(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*tes.ServerConfig, error) {
	return &taskBolt.serverConfig, nil
}

// GetJobState returns the state of a job, given a job ID.
// This is an RPC endpoint.
func (taskBolt *TaskBolt) GetJobState(ctx context.Context, id *ga4gh_task_exec.JobID) (*ga4gh_task_exec.JobDesc, error) {
	log.Debug("GetJobState called")
	var state ga4gh_task_exec.State
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		//TODO address err
		state, _ = taskBolt.getJobState(id.Value)
		return nil
	})
	jobDesc := &ga4gh_task_exec.JobDesc{
		JobID: id.Value,
		State: state,
	}
	return jobDesc, err
}
