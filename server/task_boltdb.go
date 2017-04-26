package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

// TODO these should probably be unexported names

// TaskBucket defines the name of a bucket which maps
// task ID -> tes.Task struct
var TaskBucket = []byte("tasks")

// TaskAuthBucket defines the name of a bucket which maps
// task ID -> JWT token string
var TaskAuthBucket = []byte("tasks-auth")

// TasksQueued defines the name of a bucket which maps
// task ID -> nil
var TasksQueued = []byte("tasks-queued")

// TaskState maps: task ID -> state string
var TaskState = []byte("tasks-state")

// TasksLog defines the name of a bucket which maps
// task ID -> tes.TaskLog struct
var TasksLog = []byte("tasks-log")

// Workers maps:
// worker ID -> funnel.Worker struct
var Workers = []byte("workers")

// TaskWorker Map task ID -> worker ID
var TaskWorker = []byte("task-worker")

// WorkerTasks indexes worker -> tasks
// Implemented as composite_key(worker ID + task ID) => task ID
// And searched with prefix scan using worker ID
var WorkerTasks = []byte("worker-tasks")

// TaskBolt provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type TaskBolt struct {
	db   *bolt.DB
	conf config.Config
}

// NewTaskBolt returns a new instance of TaskBolt, accessing the database at
// the given path, and including the given ServerConfig.
func NewTaskBolt(conf config.Config) (*TaskBolt, error) {
	util.EnsurePath(conf.DBPath)
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
		if tx.Bucket(TasksQueued) == nil {
			tx.CreateBucket(TasksQueued)
		}
		if tx.Bucket(TaskState) == nil {
			tx.CreateBucket(TaskState)
		}
		if tx.Bucket(TasksLog) == nil {
			tx.CreateBucket(TasksLog)
		}
		if tx.Bucket(Workers) == nil {
			tx.CreateBucket(Workers)
		}
		if tx.Bucket(TaskWorker) == nil {
			tx.CreateBucket(TaskWorker)
		}
		if tx.Bucket(WorkerTasks) == nil {
			tx.CreateBucket(WorkerTasks)
		}
		return nil
	})
	return &TaskBolt{db: db, conf: conf}, nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (taskBolt *TaskBolt) ReadQueue(n int) []*tes.Task {
	tasks := make([]*tes.Task, 0)
	taskBolt.db.View(func(tx *bolt.Tx) error {

		// Iterate over the TasksQueued bucket, reading the first `n` tasks
		c := tx.Bucket(TasksQueued).Cursor()
		for k, _ := c.First(); k != nil && len(tasks) < n; k, _ = c.Next() {
			id := string(k)
			task := getTask(tx, id)
			tasks = append(tasks, task)
		}
		return nil
	})
	return tasks
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

// GenTaskID generates a task ID string.
// IDs are globally unique and sortable.
func GenTaskID() string {
	id := xid.New()
	return id.String()
}

// CreateTask documentation
// TODO: documentation
func (taskBolt *TaskBolt) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	taskID := GenTaskID()
	log := log.WithFields("taskID", taskID)

	log.Debug("CreateTask called", "task", task)

	if len(task.Executors) == 0 {
		log.Error("No executor commands found")
		return nil, fmt.Errorf("No executor commands found")
	}

	jwt := getJWT(ctx)
	log.Debug("JWT", "token", jwt)

	ch := make(chan *tes.CreateTaskResponse, 1)
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		idBytes := []byte(taskID)

		taskopB := tx.Bucket(TaskBucket)
		v, err := proto.Marshal(task)
		if err != nil {
			return err
		}
		taskopB.Put(idBytes, v)

		tx.Bucket(TaskState).Put(idBytes, []byte(tes.State_QUEUED.String()))

		taskopA := tx.Bucket(TaskAuthBucket)
		taskopA.Put(idBytes, []byte(jwt))

		queueB := tx.Bucket(TasksQueued)
		queueB.Put(idBytes, []byte{})
		ch <- &tes.CreateTaskResponse{Id: taskID}
		return nil
	})
	if err != nil {
		log.Error("Error processing task", err)
		return nil, err
	}
	a := <-ch
	return a, err
}

func getTaskState(tx *bolt.Tx, id string) tes.State {
	idBytes := []byte(id)
	s := tx.Bucket(TaskState).Get(idBytes)
	if s == nil {
		return tes.State_UNKNOWN
	}
	// map the string into the protobuf enum
	v := tes.State_value[string(s)]
	return tes.State(v)
}

func getTask(tx *bolt.Tx, taskID string) *tes.Task {
	bT := tx.Bucket(TaskBucket)
	v := bT.Get([]byte(taskID))
	task := &tes.Task{}
	proto.Unmarshal(v, task)
	task.Id = taskID
	task.State = getTaskState(tx, taskID)
	return task
}

func loadTaskLogs(tx *bolt.Tx, task *tes.Task) {
	//if there is logging info
	bucket := tx.Bucket(TasksLog)
	out := make([]*tes.ExecutorLog, 0)

	for i := range task.Executors {
		o := bucket.Get([]byte(fmt.Sprint(task.Id, i)))
		if o != nil {
			var log tes.ExecutorLog
			proto.Unmarshal(o, &log)
			out = append(out, &log)
		}
	}

	task.Logs = []*tes.TaskLog{{
		Logs: out,
	}}
}

// GetTask gets a task, which describes a running task
func (taskBolt *TaskBolt) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {

	var task *tes.Task
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		task = getTask(tx, req.Id)
		loadTaskLogs(tx, task)
		return nil
	})
	return task, err
}

// ListTasks returns a list of taskIDs
func (taskBolt *TaskBolt) ListTasks(ctx context.Context, in *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	log.Debug("ListTasks called")

	tasks := make([]*tes.Task, 0, 10)

	taskBolt.db.View(func(tx *bolt.Tx) error {
		taskopB := tx.Bucket(TaskBucket)
		c := taskopB.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			taskID := string(k)
			taskState := getTaskState(tx, taskID)

			task := &tes.Task{}
			proto.Unmarshal(v, task)

			task = &tes.Task{
				Id:          taskID,
				State:       taskState,
				Name:        task.Name,
				Project:     task.Project,
				Description: task.Description,
			}
			tasks = append(tasks, task)
		}
		return nil
	})

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	return &out, nil
}

// CancelTask cancels a task
func (taskBolt *TaskBolt) CancelTask(ctx context.Context, taskop *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	log := log.WithFields("taskID", taskop.Id)
	log.Info("Canceling task")

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		// TODO need a test that ensures a canceled task is deleted from the worker
		id := taskop.Id
		return transitionTaskState(tx, id, tes.State_CANCELED)
	})
	if err != nil {
		return nil, err
	}
	return &tes.CancelTaskResponse{}, nil
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
	var out []string
	for _, i := range taskBolt.conf.Storage {
		if i.Local.Valid() {
			out = append(out, i.Local.AllowedDirs...)
		}
		if i.S3.Valid() {
			out = append(out, i.S3.Endpoint)
		}
	}
	return &tes.ServiceInfo{Storage: out}, nil
}
