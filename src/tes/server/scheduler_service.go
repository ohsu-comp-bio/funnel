// TODO put the boltdb implementation in a separate package
//      so that users can import pluggable backends
package tes_server

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"log"
	"tes/ga4gh"
	"tes/server/proto"
)

// GetJobToRun documentation
// TODO: documentation
func (taskBolt *TaskBolt) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResponse, error) {
	// TODO this gets really verbose. Need logging levels log.Printf("Job Request")
  var task *ga4gh_task_exec.Task
  authToken := ""

	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bQ := tx.Bucket(JobsQueued)
		bA := tx.Bucket(JobsActive)
		bOp := tx.Bucket(TaskBucket)
		authBkt := tx.Bucket(TaskAuthBucket)

		c := bQ.Cursor()

		if k, _ := c.First(); k != nil {
			log.Printf("Found queued job")

      // Get the task
			v := bOp.Get(k)
			task = &ga4gh_task_exec.Task{}
			proto.Unmarshal(v, task)
			bQ.Delete(k)

			// TODO the worker is also sending a "Running" status update, which is kind of redundant.
			//      Which is better?
      // Update the job state to "Running"
			bA.Put(k, []byte(ga4gh_task_exec.State_Running.String()))

      // Look for an auth token related to this task
      tok := authBkt.Get([]byte(task.TaskID))
      if tok != nil {
        authToken = string(tok)
      }
			return nil
		}
		return nil
	})
  // No task was found. Respond accordingly.
  if task == nil {
		return &ga4gh_task_ref.JobResponse{}, nil
  }

	job := &ga4gh_task_exec.Job{
		JobID: task.TaskID,
		Task:  task,
	}

	return &ga4gh_task_ref.JobResponse{Job: job, Auth: authToken}, nil
}

// UpdateJobStatus documentation
// TODO: documentation
func (taskBolt *TaskBolt) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobID, error) {
	log.Printf("Set job status")
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JobsActive)
		bc := tx.Bucket(JobsComplete)
		bL := tx.Bucket(JobsLog)

		if stat.Log != nil {
			log.Printf("Logging stdout:%s", stat.Log.Stdout)
			d, _ := proto.Marshal(stat.Log)
			bL.Put([]byte(fmt.Sprint(stat.Id, stat.Step)), d)
		}

		switch stat.State {
		case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error:
			ba.Delete([]byte(stat.Id))
			bc.Put([]byte(stat.Id), []byte(stat.State.String()))
		}
		return nil
	})
	return &ga4gh_task_exec.JobID{Value: stat.Id}, nil
}

// WorkerPing documentation
// TODO: documentation
func (taskBolt *TaskBolt) WorkerPing(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.WorkerInfo, error) {
	log.Printf("Worker Ping")
	return info, nil
}

// GetQueueInfo documentation
// TODO: documentation
func (taskBolt *TaskBolt) GetQueueInfo(request *ga4gh_task_ref.QueuedTaskInfoRequest, server ga4gh_task_ref.Scheduler_GetQueueInfoServer) error {
	ch := make(chan *ga4gh_task_exec.Task)
	log.Printf("Getting queue info")

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

// GetServiceInfo documentation
// TODO: documentation
func (taskBolt *TaskBolt) GetServerConfig(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.ServerConfig, error) {
	return &taskBolt.serverConfig, nil
}
