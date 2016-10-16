package tes_server

import (
	"fmt"
	"tes/server/proto"
	"tes/ga4gh"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"log"
	"strings"
)

var TASK_BUCKET = []byte("tasks")

var JOBS_QUEUED = []byte("jobs-queued")
var JOBS_ACTIVE = []byte("jobs-active")
var JOBS_COMPLETE = []byte("jobs-complete")

var JOBS_LOG = []byte("jobs-log")

type TaskBolt struct {
	db               *bolt.DB
	storage_metadata map[string]string
}

func NewTaskBolt(path string, storage_metadata map[string]string) *TaskBolt {
	db, _ := bolt.Open(path, 0600, nil)
	//Check to make sure all the required buckets have been created
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(TASK_BUCKET) == nil {
			tx.CreateBucket(TASK_BUCKET)
		}
		if tx.Bucket(JOBS_QUEUED) == nil {
			tx.CreateBucket(JOBS_QUEUED)
		}
		if tx.Bucket(JOBS_ACTIVE) == nil {
			tx.CreateBucket(JOBS_ACTIVE)
		}
		if tx.Bucket(JOBS_COMPLETE) == nil {
			tx.CreateBucket(JOBS_COMPLETE)
		}
		if tx.Bucket(JOBS_LOG) == nil {
			tx.CreateBucket(JOBS_LOG)
		}
		return nil
	})
	return &TaskBolt{db: db, storage_metadata: storage_metadata}
}

// / Run a task
func (self *TaskBolt) RunTask(ctx context.Context, task *ga4gh_task_exec.Task) (*ga4gh_task_exec.JobId, error) {
	log.Println("Recieving Task for Queue", task)

	taskopId, _ := uuid.NewV4()

	task.TaskId = taskopId.String()
	if len(task.Docker) == 0 {
		return nil, fmt.Errorf("No docker commands found")
	}

	// Check inputs of the task
	for _, input := range task.GetInputs() {
		disk_found := false
		for _, res := range task.Resources.Volumes {
			if strings.HasPrefix(input.Path, res.MountPoint) {
				disk_found = true
			}
		}
		if !disk_found {
			return nil, fmt.Errorf("Required volume '%s' not found in resources", input.Path)
		}
	}

	ch := make(chan *ga4gh_task_exec.JobId, 1)
	err := self.db.Update(func(tx *bolt.Tx) error {

		taskop_b := tx.Bucket(TASK_BUCKET)
		v, _ := proto.Marshal(task)
		taskop_b.Put([]byte(taskopId.String()), v)

		queue_b := tx.Bucket(JOBS_QUEUED)
		queue_b.Put([]byte(taskopId.String()), []byte(ga4gh_task_exec.State_Queued.String()))
		ch <- &ga4gh_task_exec.JobId{Value: taskopId.String()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := <-ch
	return a, err
}

func (self *TaskBolt) getTaskJob(task *ga4gh_task_exec.Task) (*ga4gh_task_exec.Job, error) {
	ch := make(chan *ga4gh_task_exec.Job, 1)
	self.db.View(func(tx *bolt.Tx) error {
		//right now making the assumption that the taskid is the jobid
		b_q := tx.Bucket(JOBS_QUEUED)
		b_a := tx.Bucket(JOBS_ACTIVE)
		b_c := tx.Bucket(JOBS_COMPLETE)
		b_l := tx.Bucket(JOBS_LOG)

		job := ga4gh_task_exec.Job{}
		job.Task = task
		job.JobId = task.TaskId
		//if its queued
		if v := b_q.Get([]byte(task.TaskId)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if its active
		if v := b_a.Get([]byte(task.TaskId)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if its complete
		if v := b_c.Get([]byte(job.JobId)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if there is logging info
		out := make([]*ga4gh_task_exec.JobLog, len(job.Task.Docker), len(job.Task.Docker))
		for i := range job.Task.Docker {
			o := b_l.Get([]byte(fmt.Sprint("%s-%d", job.JobId, i)))
			if o != nil {
				var log ga4gh_task_exec.JobLog
				proto.Unmarshal(o, &log)
				out[i] = &log
			} else {
				out[i] = &ga4gh_task_exec.JobLog{}
			}
		}
		job.Logs = out
		ch <- &job
		return nil
	})
	a := <-ch
	return a, nil
}

// / Get info about a running task
func (self *TaskBolt) GetJob(ctx context.Context, job *ga4gh_task_exec.JobId) (*ga4gh_task_exec.Job, error) {
	log.Printf("Getting Task Info")
	ch := make(chan *ga4gh_task_exec.Task, 1)
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		v := b.Get([]byte(job.Value))
		out := ga4gh_task_exec.Task{}
		proto.Unmarshal(v, &out)
		ch <- &out
		return nil
	})
	a := <-ch
	if a == nil {
		return nil, fmt.Errorf("Job Not Found")
	}
	b, err := self.getTaskJob(a)
	return b, err
}

func (self *TaskBolt) ListJobs(ctx context.Context, in *ga4gh_task_exec.JobListRequest) (*ga4gh_task_exec.JobListResponse, error) {
	log.Printf("Getting Task List")
	ch := make(chan *ga4gh_task_exec.Task, 1)
	go self.db.View(func(tx *bolt.Tx) error {
		taskop_b := tx.Bucket(TASK_BUCKET)
		c := taskop_b.Cursor()
		log.Println("Scanning")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			out := ga4gh_task_exec.Task{}
			proto.Unmarshal(v, &out)
			ch <- &out
		}
		close(ch)
		return nil
	})

	task_array := make([]*ga4gh_task_exec.Job, 0, 10)
	for t := range ch {
		j, _ := self.getTaskJob(t)
		task_array = append(task_array, j)
	}

	out := ga4gh_task_exec.JobListResponse{
		Jobs: task_array,
	}
	fmt.Println("Returning", out)
	return &out, nil

}

// / Cancel a running task
func (self *TaskBolt) CancelJob(ctx context.Context, taskop *ga4gh_task_exec.JobId) (*ga4gh_task_exec.JobId, error) {
	self.db.Update(func(tx *bolt.Tx) error {
		b_q := tx.Bucket(JOBS_QUEUED)
		b_q.Delete([]byte(taskop.Value))

		b_a := tx.Bucket(JOBS_ACTIVE)
		b_a.Delete([]byte(taskop.Value))

		b_c := tx.Bucket(JOBS_COMPLETE)
		b_c.Put([]byte(taskop.Value), []byte(ga4gh_task_exec.State_Canceled.String()))
		return nil
	})
	return taskop, nil
}

func (self *TaskBolt) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResponse, error) {
	//log.Printf("Job Request")
	ch := make(chan *ga4gh_task_exec.Task, 1)

	self.db.Update(func(tx *bolt.Tx) error {
		b_q := tx.Bucket(JOBS_QUEUED)
		b_a := tx.Bucket(JOBS_ACTIVE)
		b_op := tx.Bucket(TASK_BUCKET)

		c := b_q.Cursor()

		if k, _ := c.First(); k != nil {
			log.Printf("Found queued job")
			v := b_op.Get(k)
			out := ga4gh_task_exec.Task{}
			proto.Unmarshal(v, &out)
			ch <- &out
			b_q.Delete(k)
			b_a.Put(k, []byte(ga4gh_task_exec.State_Running.String()))
			return nil
		}
		ch <- nil
		return nil
	})
	a := <-ch
	if a == nil {
		return &ga4gh_task_ref.JobResponse{}, nil
	}

	job := &ga4gh_task_exec.Job{
		JobId: a.TaskId,
		Task:  a,
	}

	return &ga4gh_task_ref.JobResponse{Job: job}, nil
}

func (self *TaskBolt) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobId, error) {
	log.Printf("Set op status")
	self.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(JOBS_ACTIVE)
		bc := tx.Bucket(JOBS_COMPLETE)
		b_l := tx.Bucket(JOBS_LOG)

		if stat.Log != nil {
			log.Printf("Logging stdout:%s", stat.Log.Stdout)
			d, _ := proto.Marshal(stat.Log)
			b_l.Put([]byte(fmt.Sprint("%s-%d", stat.Id, stat.Step)), d)
		}

		switch stat.State {
		case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error:
			ba.Delete([]byte(stat.Id))
			bc.Put([]byte(stat.Id), []byte(stat.State.String()))
		}
		return nil
	})
	return &ga4gh_task_exec.JobId{Value: stat.Id}, nil
}

func (self *TaskBolt) WorkerPing(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.WorkerInfo, error) {
	log.Printf("Worker Ping")
	return info, nil
}

func (self *TaskBolt) GetServiceInfo(ctx context.Context, info *ga4gh_task_exec.ServiceInfoRequest) (*ga4gh_task_exec.ServiceInfo, error) {
	return &ga4gh_task_exec.ServiceInfo{StorageConfig: self.storage_metadata}, nil
}

func (self *TaskBolt) GetQueueInfo(request *ga4gh_task_ref.QueuedTaskInfoRequest, server ga4gh_task_ref.Scheduler_GetQueueInfoServer) error {
	ch := make(chan *ga4gh_task_exec.Task, 10)

	self.db.View(func(tx *bolt.Tx) error {
		bt := tx.Bucket(TASK_BUCKET)
		bq := tx.Bucket(JOBS_QUEUED)
		c := bq.Cursor()
		var count int32 = 0
		for k, v := c.First(); k != nil && count < request.MaxTasks; k, v = c.Next() {
			if string(v) == ga4gh_task_exec.State_Queued.String() {
				v := bt.Get(k)
				out := ga4gh_task_exec.Task{}
				proto.Unmarshal(v, &out)
				ch <- &out
			}
		}
		close(ch)
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
