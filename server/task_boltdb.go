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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// ExecutorLogs maps (task ID + executor index) -> tes.ExecutorLog struct
var ExecutorLogs = []byte("executor-logs")

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
		if tx.Bucket(ExecutorLogs) == nil {
			tx.CreateBucket(ExecutorLogs)
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

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (taskBolt *TaskBolt) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	log.Debug("CreateTask called", "task", task)

	if err := tes.Validate(task); err != nil {
		log.Error("Invalid task message", "error", err)
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	taskID := GenTaskID()
	log := log.WithFields("taskID", taskID)

	jwt := getJWT(ctx)
	log.Debug("JWT", "token", jwt)

	ch := make(chan *tes.CreateTaskResponse, 1)
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		idBytes := []byte(taskID)

		taskopB := tx.Bucket(TaskBucket)
		task.Id = taskID
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

func loadBasicTaskView(tx *bolt.Tx, id string, task *tes.Task) {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	proto.Unmarshal(b, task)
}

func loadTaskLogs(tx *bolt.Tx, task *tes.Task) {
	tasklog := &tes.TaskLog{}
	task.Logs = []*tes.TaskLog{tasklog}

	b := tx.Bucket(TasksLog).Get([]byte(task.Id))
	if b != nil {
		proto.Unmarshal(b, tasklog)
	}

	for i := range task.Executors {
		o := tx.Bucket(ExecutorLogs).Get([]byte(fmt.Sprint(task.Id, i)))
		if o != nil {
			var execlog tes.ExecutorLog
			proto.Unmarshal(o, &execlog)
			tasklog.Logs = append(tasklog.Logs, &execlog)
		}
	}
}

// GetTask gets a task, which describes a running task
func (taskBolt *TaskBolt) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		task = getTaskView(tx, req.Id, req.View)
		return nil
	})
	return task, err
}

func getTask(tx *bolt.Tx, id string) *tes.Task {
	// This is a thin wrapper around getTaskView in order to allow task views
	// to be added with out changing existing code calling getTask().
	return getTaskView(tx, id, tes.TaskView_FULL)
}

func getTaskView(tx *bolt.Tx, id string, view tes.TaskView) *tes.Task {
	task := &tes.Task{}

	if view == tes.TaskView_BASIC {
		loadBasicTaskView(tx, id, task)
	} else if view == tes.TaskView_FULL {
		loadBasicTaskView(tx, id, task)
		loadTaskLogs(tx, task)
	}
	task.Id = id
	task.State = getTaskState(tx, id)

	return task
}

// ListTasks returns a list of taskIDs
func (taskBolt *TaskBolt) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	var tasks []*tes.Task
	pageSize := 256

	if req.PageSize != 0 {
		pageSize = int(req.GetPageSize())
		if pageSize > 2048 {
			pageSize = 2048
		}
		if pageSize < 50 {
			pageSize = 50
		}
	}

	taskBolt.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(TaskBucket).Cursor()

		i := 0

		// For pagination, figure out the starting key.
		var k []byte
		if req.PageToken != "" {
			// Seek moves to the key, but the start of the page is the next key.
			c.Seek([]byte(req.PageToken))
			k, _ = c.Next()
		} else {
			// No pagination, so take the first key.
			k, _ = c.First()
		}

		for ; k != nil && i < pageSize; k, _ = c.Next() {
			task := getTaskView(tx, string(k), req.View)
			tasks = append(tasks, task)
			i++
		}
		return nil
	})

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	if len(tasks) == pageSize {
		out.NextPageToken = tasks[len(tasks)-1].Id
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
	// BUG: this isn't the best translation, probably lossy.
	//     Maybe ServiceInfo data structure schema needs to be refactored
	//     For example, you can't have multiple S3 endpoints
	var out []string
	if taskBolt.conf.Storage.Local.Valid() {
		out = append(out, taskBolt.conf.Storage.Local.AllowedDirs...)
	}

	for _, i := range taskBolt.conf.Storage.S3 {
		if i.Valid() {
			out = append(out, i.Endpoint)
		}
	}
	return &tes.ServiceInfo{Name: taskBolt.conf.ServiceName, Storage: out}, nil
}
