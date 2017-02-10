package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"strings"
	"tes/config"
	"tes/ga4gh"
	"time"
)

// TODO these should probably be unexported names

// TaskBucket defines the name of a bucket which maps
// job ID -> ga4gh_task_exec.Task struct
var TaskBucket = []byte("tasks")

// TaskAuthBucket defines the name of a bucket which maps
// job ID -> JWT token string
var TaskAuthBucket = []byte("tasks-auth")

// JobsQueued defines the name of a bucket which maps
// job ID -> nil
var JobsQueued = []byte("jobs-queued")

// JobState maps: job ID -> state string
var JobState = []byte("jobs-state")

// JobsLog defines the name of a bucket which maps
// job ID -> ga4gh_task_exec.JobLog struct
var JobsLog = []byte("jobs-log")

// JobWorker defines the name a bucket which maps
// job ID -> worker ID
var JobWorker = []byte("job-worker")

// Workers maps:
// worker ID -> ga4gh_task_ref.Worker struct
var Workers = []byte("workers")

// TaskBolt provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type TaskBolt struct {
	db           *bolt.DB
	serverConfig config.Config
}

// NewTaskBolt returns a new instance of TaskBolt, accessing the database at
// the given path, and including the given ServerConfig.
func NewTaskBolt(conf config.Config) (*TaskBolt, error) {
	db, err := bolt.Open(conf.DBPath, 0600, &bolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}

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
		if tx.Bucket(JobState) == nil {
			tx.CreateBucket(JobState)
		}
		if tx.Bucket(JobsLog) == nil {
			tx.CreateBucket(JobsLog)
		}
		if tx.Bucket(Workers) == nil {
			tx.CreateBucket(Workers)
		}
		if tx.Bucket(JobWorker) == nil {
			tx.CreateBucket(JobWorker)
		}
		return nil
	})
	return &TaskBolt{db: db, serverConfig: conf}, nil
}

// ReadQueue returns a slice of queued Jobs. Up to "n" jobs are returned.
func (taskBolt *TaskBolt) ReadQueue(n int) []*ga4gh_task_exec.Job {
	jobs := make([]*ga4gh_task_exec.Job, 0)
	taskBolt.db.View(func(tx *bolt.Tx) error {

		// Iterate over the JobsQueued bucket, reading the first `n` jobs
		c := tx.Bucket(JobsQueued).Cursor()
		for k, _ := c.First(); k != nil && len(jobs) < n; k, _ = c.Next() {
			id := string(k)
			job := getJob(tx, id)
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
	jobID, _ := uuid.NewV4()
	log := log.WithFields("jobID", jobID)

	log.Debug("RunTask called", "task", task)

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
	log.Debug("JWT", "token", jwt)

	ch := make(chan *ga4gh_task_exec.JobID, 1)
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		taskopB := tx.Bucket(TaskBucket)
		v, _ := proto.Marshal(task)
		taskopB.Put([]byte(jobID.String()), v)

		taskopA := tx.Bucket(TaskAuthBucket)
		taskopA.Put([]byte(jobID.String()), []byte(jwt))

		queueB := tx.Bucket(JobsQueued)
		queueB.Put([]byte(jobID.String()), []byte{})
		ch <- &ga4gh_task_exec.JobID{Value: jobID.String()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := <-ch
	return a, err
}

func getJobState(tx *bolt.Tx, id string) ga4gh_task_exec.State {
	idBytes := []byte(id)
	s := tx.Bucket(JobState).Get(idBytes)
	if s == nil {
		return ga4gh_task_exec.State_Unknown
	} else {
		// map the string into the protobuf enum
		v := ga4gh_task_exec.State_value[string(s)]
		return ga4gh_task_exec.State(v)
	}
}

func getJob(tx *bolt.Tx, jobID string) *ga4gh_task_exec.Job {
	bT := tx.Bucket(TaskBucket)
	v := bT.Get([]byte(jobID))
	task := &ga4gh_task_exec.Task{}
	proto.Unmarshal(v, task)

	job := ga4gh_task_exec.Job{}
	job.JobID = jobID
	job.Task = task
	job.State = getJobState(tx, jobID)
	return &job
}

func loadJobLogs(tx *bolt.Tx, job *ga4gh_task_exec.Job) {
	//if there is logging info
	bucket := tx.Bucket(JobsLog)
	out := make([]*ga4gh_task_exec.JobLog, 0)

	for i := range job.Task.Docker {
		o := bucket.Get([]byte(fmt.Sprint(job.JobID, i)))
		if o != nil {
			var log ga4gh_task_exec.JobLog
			proto.Unmarshal(o, &log)
			out = append(out, &log)
		}
	}

	job.Logs = out
}

// GetJob gets a job, which describes a running task
func (taskBolt *TaskBolt) GetJob(ctx context.Context, id *ga4gh_task_exec.JobID) (*ga4gh_task_exec.Job, error) {
	log.Debug("GetJob called", "jobID", id.Value)

	var job *ga4gh_task_exec.Job
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		job = getJob(tx, id.Value)
		loadJobLogs(tx, job)
		return nil
	})
	return job, err
}

// ListJobs returns a list of jobIDs
func (taskBolt *TaskBolt) ListJobs(ctx context.Context, in *ga4gh_task_exec.JobListRequest) (*ga4gh_task_exec.JobListResponse, error) {
	log.Debug("ListJobs called")

	jobs := make([]*ga4gh_task_exec.JobDesc, 0, 10)

	taskBolt.db.View(func(tx *bolt.Tx) error {
		taskopB := tx.Bucket(TaskBucket)
		c := taskopB.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			jobID := string(k)
			jobState := getJobState(tx, jobID)

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

	return &out, nil
}

// CancelJob cancels a job
func (taskBolt *TaskBolt) CancelJob(ctx context.Context, taskop *ga4gh_task_exec.JobID) (*ga4gh_task_exec.JobID, error) {
	log := log.WithFields("jobID", taskop.Value)
	log.Info("Canceling job")

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		return transitionJobState(tx, taskop.Value, ga4gh_task_exec.State_Canceled)
	})
	if err != nil {
		return nil, err
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
		if i.Local.Valid() {
			out["Local.AllowedDirs"] = strings.Join(i.Local.AllowedDirs, ",")
		}
		if i.S3.Valid() {
			out["S3.Endpoint"] = i.S3.Endpoint
		}
	}
	return &ga4gh_task_exec.ServiceInfo{StorageConfig: out}, nil
}
