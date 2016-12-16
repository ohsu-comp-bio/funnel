package tes_server

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
	"tes/ga4gh"
	"tes/server/proto"
)

// TaskBucket documentation
// TODO: documentation
var TaskBucket = []byte("tasks")

// TaskAuthBucket documentation
// TODO: documentation
var TaskAuthBucket = []byte("tasks-auth")

// JobsQueued documentation
// TODO: documentation
var JobsQueued = []byte("jobs-queued")

// JobsActive documentation
// TODO: documentation
var JobsActive = []byte("jobs-active")

// JobsComplete documentation
// TODO: documentation
var JobsComplete = []byte("jobs-complete")

// JobsLog documentation
// TODO: documentation
var JobsLog = []byte("jobs-log")

// TaskBolt documentation
// TODO: documentation
type TaskBolt struct {
	db           *bolt.DB
	serverConfig ga4gh_task_ref.ServerConfig
}

// NewTaskBolt documentation
// TODO: documentation
func NewTaskBolt(path string, config ga4gh_task_ref.ServerConfig) *TaskBolt {
	db, _ := bolt.Open(path, 0600, nil)
	//Check to make sure all the required buckets have been created
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(TaskBucket) == nil {
			tx.CreateBucket(TaskBucket)
		}
		if tx.Bucket(TaskAuthBucket) == nil {
			tx.CreateBucket(TaskAuthBucket)
		}
		if tx.Bucket(JobsQueued) == nil {
			tx.CreateBucket(JobsQueued)
		}
		if tx.Bucket(JobsActive) == nil {
			tx.CreateBucket(JobsActive)
		}
		if tx.Bucket(JobsComplete) == nil {
			tx.CreateBucket(JobsComplete)
		}
		if tx.Bucket(JobsLog) == nil {
			tx.CreateBucket(JobsLog)
		}
		return nil
	})
	return &TaskBolt{db: db, serverConfig: config}
}

// getJWT
// This function extracts the JWT token from the rpc header and returns the string
func getJWT(ctx context.Context) string {
	jwt := ""
	v, _ := metadata.FromContext(ctx)
	auth, ok := v["authorization"]
	if !ok {
		return jwt
	}
	for _, i := range auth {
		if strings.HasPrefix(i, "JWT ") {
			jwt = strings.TrimPrefix(i, "JWT ")
		}
	}
	return jwt
}

// RunTask documentation
// TODO: documentation
func (taskBolt *TaskBolt) RunTask(ctx context.Context, task *ga4gh_task_exec.Task) (*ga4gh_task_exec.JobID, error) {
	log.Println("Receiving Task for Queue", task)

	taskopID, _ := uuid.NewV4()

	task.TaskID = taskopID.String()
	if len(task.Docker) == 0 {
		return nil, fmt.Errorf("No docker commands found")
	}

	// Check inputs of the task
	for _, input := range task.GetInputs() {
		diskFound := false
		for _, res := range task.Resources.Volumes {
			if strings.HasPrefix(input.Path, res.MountPoint) {
				diskFound = true
			}
		}
		if !diskFound {
			return nil, fmt.Errorf("Required volume '%s' not found in resources", input.Path)
		}
		//Fixing blank value to File by default... Is this too much hand holding?
		if input.Class == "" {
			input.Class = "File"
		}
	}

	for _, output := range task.GetOutputs() {
		if output.Class == "" {
			output.Class = "File"
		}
	}

	jwt := getJWT(ctx)
	log.Printf("JWT: %s", jwt)

	ch := make(chan *ga4gh_task_exec.JobID, 1)
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		taskopB := tx.Bucket(TaskBucket)
		v, _ := proto.Marshal(task)
		taskopB.Put([]byte(taskopID.String()), v)

		taskopA := tx.Bucket(TaskAuthBucket)
		taskopA.Put([]byte(taskopID.String()), []byte(jwt))

		queueB := tx.Bucket(JobsQueued)
		queueB.Put([]byte(taskopID.String()), []byte(ga4gh_task_exec.State_Queued.String()))
		ch <- &ga4gh_task_exec.JobID{Value: taskopID.String()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := <-ch
	return a, err
}

func (taskBolt *TaskBolt) getTaskJob(task *ga4gh_task_exec.Task) (*ga4gh_task_exec.Job, error) {
	ch := make(chan *ga4gh_task_exec.Job, 1)
	taskBolt.db.View(func(tx *bolt.Tx) error {
		//right now making the assumption that the taskid is the jobid
		bQ := tx.Bucket(JobsQueued)
		bA := tx.Bucket(JobsActive)
		bC := tx.Bucket(JobsComplete)
		bL := tx.Bucket(JobsLog)

		job := ga4gh_task_exec.Job{}
		job.Task = task
		job.JobID = task.TaskID
		//if its queued
		if v := bQ.Get([]byte(task.TaskID)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if its active
		if v := bA.Get([]byte(task.TaskID)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if its complete
		if v := bC.Get([]byte(job.JobID)); v != nil {
			job.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if there is logging info
		out := make([]*ga4gh_task_exec.JobLog, len(job.Task.Docker), len(job.Task.Docker))
		for i := range job.Task.Docker {
			o := bL.Get([]byte(fmt.Sprint(job.JobID, i)))
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

// GetJob documentation
// TODO: documentation
// Get info about a running task
func (taskBolt *TaskBolt) GetJob(ctx context.Context, job *ga4gh_task_exec.JobID) (*ga4gh_task_exec.Job, error) {
	log.Printf("Getting Task Info")
	ch := make(chan *ga4gh_task_exec.Task, 1)
	taskBolt.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TaskBucket)
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
	b, err := taskBolt.getTaskJob(a)
	return b, err
}

// ListJobs returns a list of jobIDs
func (taskBolt *TaskBolt) ListJobs(ctx context.Context, in *ga4gh_task_exec.JobListRequest) (*ga4gh_task_exec.JobListResponse, error) {
	log.Printf("Getting Task List")
	ch := make(chan *ga4gh_task_exec.Task, 1)
	go taskBolt.db.View(func(tx *bolt.Tx) error {
		taskopB := tx.Bucket(TaskBucket)
		c := taskopB.Cursor()
		log.Println("Scanning")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			out := ga4gh_task_exec.Task{}
			proto.Unmarshal(v, &out)
			ch <- &out
		}
		close(ch)
		return nil
	})

	jobIDArray := make([]string, 0, 10)
	jobStateArray := make([]ga4gh_task_exec.State, 0, 10)

	for t := range ch {
		j, _ := taskBolt.getTaskJob(t)
		jobIDArray = append(jobIDArray, j.JobID)
		jobStateArray = append(jobStateArray, j.State)
	}

	out := ga4gh_task_exec.JobListResponse{
		JobID:    jobIDArray,
		State: jobStateArray,
	}
	fmt.Println("Returning", out)
	return &out, nil

}

// CancelJob documentation
// TODO: documentation
// Cancel a running task
func (taskBolt *TaskBolt) CancelJob(ctx context.Context, taskop *ga4gh_task_exec.JobID) (*ga4gh_task_exec.JobID, error) {
	taskBolt.db.Update(func(tx *bolt.Tx) error {
		bQ := tx.Bucket(JobsQueued)
		bQ.Delete([]byte(taskop.Value))

		bA := tx.Bucket(JobsActive)
		bA.Delete([]byte(taskop.Value))

		bC := tx.Bucket(JobsComplete)
		bC.Put([]byte(taskop.Value), []byte(ga4gh_task_exec.State_Canceled.String()))
		return nil
	})
	return taskop, nil
}

// GetServiceInfo documentation
// TODO: documentation
func (taskBolt *TaskBolt) GetServiceInfo(ctx context.Context, info *ga4gh_task_exec.ServiceInfoRequest) (*ga4gh_task_exec.ServiceInfo, error) {
	//BUG: this isn't the best translation, probably lossy. Maybe ServiceInfo data structure schema needs to be refactored
	out := map[string]string{}
	for _, i := range taskBolt.serverConfig.Storage {
		for j, k := range i.Config {
			out[fmt.Sprintf("%s.%s", i.Protocol, j)] = k
		}
	}
	return &ga4gh_task_exec.ServiceInfo{StorageConfig: out}, nil
}
