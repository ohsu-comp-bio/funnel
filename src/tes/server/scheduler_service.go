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
	var task *ga4gh_task_exec.Task
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
			proto.Unmarshal(v, task)

			// Look for an auth token related to this task
			tok := authBkt.Get([]byte(task.TaskID))
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
		JobID: task.TaskID,
		Task:  task,
	}

	return &ga4gh_task_ref.JobResponse{Job: job, Auth: authToken}, nil
}

func (taskBolt *TaskBolt) AssignTask(id string, workerID string) error {
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JobsActive)
		bc := tx.Bucket(JobsComplete)
    bq := tx.Bucket(JobsQueued)
    bw := tx.Bucket(WorkerJobs)
    k := []byte(id)
    w := []byte(workerID)
    bc.Delete(k)
    bq.Delete(k)
    ba.Put(k, w)
    bw.Put(w, k)
    return nil
  })
  return nil
}

// UpdateJobStatus documentation
// TODO: documentation
func (taskBolt *TaskBolt) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobID, error) {
	log.Printf("Set job status")
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JobsActive)
		bc := tx.Bucket(JobsComplete)
		bL := tx.Bucket(JobsLog)
    bw := tx.Bucket(WorkerJobs)

		if stat.Log != nil {
			log.Printf("Logging stdout:%s", stat.Log.Stdout)
			d, _ := proto.Marshal(stat.Log)
			bL.Put([]byte(fmt.Sprint(stat.Id, stat.Step)), d)
		}

		switch stat.State {
		case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error:
      workerID := ba.Get([]byte(stat.Id))
			ba.Delete([]byte(stat.Id))
      bw.Delete([]byte(workerID))
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
