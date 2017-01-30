package server

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

// TODO these should probably be unexported names

// TaskBucket defines the name of a bucket which maps
// job ID -> ga4gh_task_exec.Task struct
var TaskBucket = []byte("tasks")

// TaskAuthBucket defines the name of a bucket which maps
// job ID -> JWT token string
var TaskAuthBucket = []byte("tasks-auth")

// JobsQueued defines the name of a bucket which maps
// job ID -> job state string
var JobsQueued = []byte("jobs-queued")

// JobsActive defines the name of a bucket which maps
// job ID -> job state string
var JobsActive = []byte("jobs-active")

// JobsComplete defines the name of a bucket which maps
// job ID -> job state string
var JobsComplete = []byte("jobs-complete")

// JobsLog defines the name of a bucket which maps
// job ID -> ga4gh_task_exec.JobLog struct
var JobsLog = []byte("jobs-log")

// WorkerJobs defines the name of a bucket which maps
// worker ID -> job ID
var WorkerJobs = []byte("worker-jobs")

// JobWorker defines the name a bucket which maps
// job ID -> worker ID
var JobWorker = []byte("job-worker")

// TaskBolt provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type TaskBolt struct {
	db           *bolt.DB
	serverConfig ga4gh_task_ref.ServerConfig
}

// NewTaskBolt returns a new instance of TaskBolt, accessing the database at
// the given path, and including the given ServerConfig.
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
		if tx.Bucket(WorkerJobs) == nil {
			tx.CreateBucket(WorkerJobs)
		}
		if tx.Bucket(JobWorker) == nil {
			tx.CreateBucket(JobWorker)
		}
		return nil
	})
	return &TaskBolt{db: db, serverConfig: config}
}

// ReadQueue returns a slice of queued Jobs. Up to "n" jobs are returned.
func (taskBolt *TaskBolt) ReadQueue(n int) []*ga4gh_task_exec.Job {
	jobs := make([]*ga4gh_task_exec.Job, 0)
	taskBolt.db.View(func(tx *bolt.Tx) error {

		// Iterate over the JobsQueued bucket, reading the first `n` jobs
		c := tx.Bucket(JobsQueued).Cursor()
		for k, _ := c.First(); k != nil && len(jobs) < n; k, _ = c.Next() {
			id := string(k)
			job := taskBolt.getJob(tx, id)
			jobs = append(jobs, job)
		}
		return nil
	})
	return jobs
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

	jobID, _ := uuid.NewV4()
	log.Printf("Assigning job ID, %s", jobID)

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
		taskopB.Put([]byte(jobID.String()), v)

		taskopA := tx.Bucket(TaskAuthBucket)
		taskopA.Put([]byte(jobID.String()), []byte(jwt))

		queueB := tx.Bucket(JobsQueued)
		queueB.Put([]byte(jobID.String()), []byte(ga4gh_task_exec.State_Queued.String()))
		ch <- &ga4gh_task_exec.JobID{Value: jobID.String()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := <-ch
	return a, err
}

func (taskBolt *TaskBolt) getJobState(jobID string) (ga4gh_task_exec.State, error) {

	ch := make(chan ga4gh_task_exec.State, 1)
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		bQ := tx.Bucket(JobsQueued)
		bA := tx.Bucket(JobsActive)
		bC := tx.Bucket(JobsComplete)

		if v := bQ.Get([]byte(jobID)); v != nil {
			//if its queued
			ch <- ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		} else if v := bA.Get([]byte(jobID)); v != nil {
			//if its active
			ch <- ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		} else if v := bC.Get([]byte(jobID)); v != nil {
			//if its complete
			ch <- ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		} else {
			ch <- ga4gh_task_exec.State_Unknown
		}
		return nil
	})
	a := <-ch
	return a, err
}

func (taskBolt *TaskBolt) getJob(tx *bolt.Tx, jobID string) *ga4gh_task_exec.Job {
	bT := tx.Bucket(TaskBucket)
	v := bT.Get([]byte(jobID))
	task := &ga4gh_task_exec.Task{}
	proto.Unmarshal(v, task)

	job := ga4gh_task_exec.Job{}
	job.JobID = jobID
	job.Task = task
	job.State, _ = taskBolt.getJobState(jobID)

	//if there is logging info
	bL := tx.Bucket(JobsLog)
	out := make([]*ga4gh_task_exec.JobLog, 0)

	for i := range job.Task.Docker {
		o := bL.Get([]byte(fmt.Sprint(jobID, i)))
		if o != nil {
			var log ga4gh_task_exec.JobLog
			proto.Unmarshal(o, &log)
			out = append(out, &log)
		}
	}

	job.Logs = out
	return &job
}

// GetJob documentation
// TODO: documentation
// Get info about a running task
func (taskBolt *TaskBolt) GetJob(ctx context.Context, id *ga4gh_task_exec.JobID) (*ga4gh_task_exec.Job, error) {
	log.Printf("Getting Task Info")
	var job *ga4gh_task_exec.Job
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		job = taskBolt.getJob(tx, id.Value)
		return nil
	})
	return job, err
}

// ListJobs returns a list of jobIDs
func (taskBolt *TaskBolt) ListJobs(ctx context.Context, in *ga4gh_task_exec.JobListRequest) (*ga4gh_task_exec.JobListResponse, error) {
	log.Printf("Getting Task List")

	jobs := make([]*ga4gh_task_exec.JobDesc, 0, 10)

	taskBolt.db.View(func(tx *bolt.Tx) error {
		taskopB := tx.Bucket(TaskBucket)
		c := taskopB.Cursor()
		log.Println("Scanning")

		for k, v := c.First(); k != nil; k, v = c.Next() {
			jobID := string(k)
			jobState, _ := taskBolt.getJobState(jobID)

			task := &ga4gh_task_exec.Task{}
			proto.Unmarshal(v, task)

			job := &ga4gh_task_exec.JobDesc{
				JobID: jobID,
				State: jobState,
				Task: &ga4gh_task_exec.TaskDesc{
					Name:        task.Name,
					ProjectID:   task.ProjectID,
					Description: task.Description,
				},
			}
			jobs = append(jobs, job)
		}
		return nil
	})

	out := ga4gh_task_exec.JobListResponse{
		Jobs: jobs,
	}

	log.Println("Returning", out)
	return &out, nil
}

// CancelJob documentation
// TODO: documentation
// Cancel a running task
func (taskBolt *TaskBolt) CancelJob(ctx context.Context, taskop *ga4gh_task_exec.JobID) (*ga4gh_task_exec.JobID, error) {
	state, _ := taskBolt.getJobState(taskop.Value)
	switch state {
	case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error, ga4gh_task_exec.State_Canceled:
		log.Printf("Cannot cancel a job already in a terminal status: %s", taskop.Value)
		return taskop, nil
	default:
		log.Printf("Cancelling job: %s", taskop.Value)
		taskBolt.db.Update(func(tx *bolt.Tx) error {
			bQ := tx.Bucket(JobsQueued)
			bQ.Delete([]byte(taskop.Value))
			bjw := tx.Bucket(JobWorker)

			bA := tx.Bucket(JobsActive)
			bA.Delete([]byte(taskop.Value))

			workerID := bjw.Get([]byte(taskop.Value))
			bjw.Delete([]byte(taskop.Value))

			bW := tx.Bucket(WorkerJobs)
			bW.Delete([]byte(workerID))

			bC := tx.Bucket(JobsComplete)
			bC.Put([]byte(taskop.Value), []byte(ga4gh_task_exec.State_Canceled.String()))
			return nil
		})
	}
	return taskop, nil
}

// GetServiceInfo provides an endpoint for TES clients to get information about this server.
// Could include:
// - resource availability
// - support storage systems
// - versions
// - etc.
func (taskBolt *TaskBolt) GetServiceInfo(ctx context.Context, info *ga4gh_task_exec.ServiceInfoRequest) (*ga4gh_task_exec.ServiceInfo, error) {
	//BUG: this isn't the best translation, probably lossy.
	//     Maybe ServiceInfo data structure schema needs to be refactored
	//     For example, you can't have multiple S3 endpoints
	out := map[string]string{}
	for _, i := range taskBolt.serverConfig.Storage {
		if i.Local != nil {
			out["Local.AllowedDirs"] = strings.Join(i.Local.AllowedDirs, ",")
		}

		if i.S3 != nil {
			out["S3.Endpoint"] = i.S3.Endpoint
		}
	}
	return &ga4gh_task_exec.ServiceInfo{StorageConfig: out}, nil
}
