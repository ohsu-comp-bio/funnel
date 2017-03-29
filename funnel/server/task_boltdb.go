package server

import (
	"fmt"
	"funnel/config"
	"funnel/proto/tes"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

// TODO these should probably be unexported names

// TaskBucket defines the name of a bucket which maps
// job ID -> tes.Task struct
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
// job ID -> tes.JobLog struct
var JobsLog = []byte("jobs-log")

// Workers maps:
// worker ID -> funnel.Worker struct
var Workers = []byte("workers")

// JobWorker Map job ID -> worker ID
var JobWorker = []byte("job-worker")

// WorkerJobs indexes worker -> jobs
// Implemented as composite_key(worker ID + job ID) => job ID
// And searched with prefix scan using worker ID
var WorkerJobs = []byte("worker-jobs")

// TaskBolt provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type TaskBolt struct {
	db   *bolt.DB
	conf config.Config
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
		if tx.Bucket(WorkerJobs) == nil {
			tx.CreateBucket(WorkerJobs)
		}
		return nil
	})
	return &TaskBolt{db: db, conf: conf}, nil
}

// ReadQueue returns a slice of queued Jobs. Up to "n" jobs are returned.
func (taskBolt *TaskBolt) ReadQueue(n int) []*tes.Job {
	jobs := make([]*tes.Job, 0)
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

// GenJobID generates a job ID string.
// IDs are globally unique and sortable.
func GenJobID() string {
	id := xid.New()
	return id.String()
}

// RunTask documentation
// TODO: documentation
func (taskBolt *TaskBolt) RunTask(ctx context.Context, task *tes.Task) (*tes.JobID, error) {
	jobID := GenJobID()
	log := log.WithFields("jobID", jobID)

	log.Debug("RunTask called", "task", task)

	if len(task.Docker) == 0 {
		log.Error("No docker commands found")
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
			log.Error("RunTask: required volume not found in resources",
				"path", input.Path)
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

	ch := make(chan *tes.JobID, 1)
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		idBytes := []byte(jobID)

		taskopB := tx.Bucket(TaskBucket)
		v, err := proto.Marshal(task)
		if err != nil {
			return err
		}
		taskopB.Put(idBytes, v)

		tx.Bucket(JobState).Put(idBytes, []byte(tes.State_Queued.String()))

		taskopA := tx.Bucket(TaskAuthBucket)
		taskopA.Put(idBytes, []byte(jwt))

		queueB := tx.Bucket(JobsQueued)
		queueB.Put(idBytes, []byte{})
		ch <- &tes.JobID{Value: jobID}
		return nil
	})
	if err != nil {
		log.Error("Error processing task", err)
		return nil, err
	}
	a := <-ch
	return a, err
}

func getJobState(tx *bolt.Tx, id string) tes.State {
	idBytes := []byte(id)
	s := tx.Bucket(JobState).Get(idBytes)
	if s == nil {
		return tes.State_Unknown
	}
	// map the string into the protobuf enum
	v := tes.State_value[string(s)]
	return tes.State(v)
}

func getJob(tx *bolt.Tx, jobID string) *tes.Job {
	bT := tx.Bucket(TaskBucket)
	v := bT.Get([]byte(jobID))
	task := &tes.Task{}
	proto.Unmarshal(v, task)

	job := tes.Job{}
	job.JobID = jobID
	job.Task = task
	job.State = getJobState(tx, jobID)
	return &job
}

func loadJobLogs(tx *bolt.Tx, job *tes.Job) {
	//if there is logging info
	bucket := tx.Bucket(JobsLog)
	out := make([]*tes.JobLog, 0)

	for i := range job.Task.Docker {
		o := bucket.Get([]byte(fmt.Sprint(job.JobID, i)))
		if o != nil {
			var log tes.JobLog
			proto.Unmarshal(o, &log)
			out = append(out, &log)
		}
	}

	job.Logs = out
}

// GetJob gets a job, which describes a running task
func (taskBolt *TaskBolt) GetJob(ctx context.Context, id *tes.JobID) (*tes.Job, error) {

	var job *tes.Job
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		job = getJob(tx, id.Value)
		loadJobLogs(tx, job)
		return nil
	})
	return job, err
}

// ListJobs returns a list of jobIDs
func (taskBolt *TaskBolt) ListJobs(ctx context.Context, in *tes.JobListRequest) (*tes.JobListResponse, error) {
	log.Debug("ListJobs called")

	jobs := make([]*tes.JobDesc, 0, 10)

	taskBolt.db.View(func(tx *bolt.Tx) error {
		taskopB := tx.Bucket(TaskBucket)
		c := taskopB.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			jobID := string(k)
			jobState := getJobState(tx, jobID)

			task := &tes.Task{}
			proto.Unmarshal(v, task)

			job := &tes.JobDesc{
				JobID: jobID,
				State: jobState,
			}
			jobs = append(jobs, job)
		}
		return nil
	})

	out := tes.JobListResponse{
		Jobs: jobs,
	}

	return &out, nil
}

// CancelJob cancels a job
func (taskBolt *TaskBolt) CancelJob(ctx context.Context, taskop *tes.JobID) (*tes.JobID, error) {
	log := log.WithFields("jobID", taskop.Value)
	log.Info("Canceling job")

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		// TODO need a test that ensures a canceled job is deleted from the worker
		id := taskop.Value
		return transitionJobState(tx, id, tes.State_Canceled)
	})
	if err != nil {
		return nil, err
	}
	return taskop, nil
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
// Could include:
// - resource availability
// - support storage systems
// - versions
// - etc.
func (taskBolt *TaskBolt) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	//BUG: this isn't the best translation, probably lossy.
	//     Maybe ServiceInfo data structure schema needs to be refactored
	//     For example, you can't have multiple S3 endpoints
	out := map[string]string{}
	for _, i := range taskBolt.conf.Storage {
		if i.Local.Valid() {
			out["Local.AllowedDirs"] = strings.Join(i.Local.AllowedDirs, ",")
		}
		if i.S3.Valid() {
			out["S3.Endpoint"] = i.S3.Endpoint
		}
	}
	return &tes.ServiceInfo{StorageConfig: out}, nil
}
