package boltdb

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"time"
)

// TODO these should probably be unexported names

// TaskBucket defines the name of a bucket which maps
// task ID -> tes.Task struct
var TaskBucket = []byte("tasks")

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

// Nodes maps:
// node ID -> pbs.Node struct
var Nodes = []byte("nodes")

// BoltDB provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type BoltDB struct {
	db      *bolt.DB
	conf    config.Config
	backend compute.Backend
}

// NewBoltDB returns a new instance of BoltDB, accessing the database at
// the given path, and including the given ServerConfig.
func NewBoltDB(conf config.Config) (*BoltDB, error) {
	util.EnsurePath(conf.Server.Databases.BoltDB.Path)
	db, err := bolt.Open(conf.Server.Databases.BoltDB.Path, 0600, &bolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}

	// Check to make sure all the required buckets have been created
	db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(TaskBucket) == nil {
			tx.CreateBucket(TaskBucket)
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
		if tx.Bucket(Nodes) == nil {
			tx.CreateBucket(Nodes)
		}
		return nil
	})
	return &BoltDB{db: db, conf: conf, backend: nil}, nil
}

// WithComputeBackend configures the BoltDB instance to use the given
// compute.Backend. The compute backend is responsible for dispatching tasks to
// schedulers / compute resources with its Submit method.
func (taskBolt *BoltDB) WithComputeBackend(backend compute.Backend) {
	taskBolt.backend = backend
}

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (taskBolt *BoltDB) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	log.Debug("CreateTask called", "task", task)

	if err := tes.Validate(task); err != nil {
		log.Error("Invalid task message", "error", err)
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	taskID := util.GenTaskID()
	idBytes := []byte(taskID)
	log := log.WithFields("taskID", taskID)

	task.Id = taskID
	taskString, err := proto.Marshal(task)
	if err != nil {
		return nil, err
	}

	err = taskBolt.db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(TaskBucket).Put(idBytes, taskString)
		tx.Bucket(TaskState).Put(idBytes, []byte(tes.State_QUEUED.String()))
		return nil
	})
	if err != nil {
		log.Error("Error storing task in database", err)
		return nil, err
	}

	err = taskBolt.backend.Submit(task)
	if err != nil {
		log.Error("Error submitting task to compute backend", err)
		derr := taskBolt.db.Update(func(tx *bolt.Tx) error {
			tx.Bucket(TaskBucket).Delete(idBytes)
			tx.Bucket(TaskState).Delete(idBytes)
			return nil
		})
		if derr != nil {
			log.Error("Error storing task in database", err)
			err = fmt.Errorf("%v\n%v", err, derr)
		}
		return nil, err
	}

	return &tes.CreateTaskResponse{Id: taskID}, nil
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

func loadMinimalTaskView(tx *bolt.Tx, id string, task *tes.Task) error {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	if b == nil {
		return errNotFound
	}
	task.Id = id
	task.State = getTaskState(tx, id)
	return nil
}

func loadBasicTaskView(tx *bolt.Tx, id string, task *tes.Task) error {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	if b == nil {
		return errNotFound
	}
	proto.Unmarshal(b, task)
	loadTaskLogs(tx, task)

	// remove contents from inputs
	inputs := []*tes.TaskParameter{}
	for _, v := range task.Inputs {
		v.Contents = ""
		inputs = append(inputs, v)
	}
	task.Inputs = inputs

	// remove stdout and stderr from Task.Logs.Logs
	tlogs := []*tes.TaskLog{}
	for _, tl := range task.Logs {
		elogs := []*tes.ExecutorLog{}
		for _, el := range tl.Logs {
			el.Stdout = ""
			el.Stderr = ""
			elogs = append(elogs, el)
		}
		tl.Logs = elogs
		tlogs = append(tlogs, tl)
	}
	task.Logs = tlogs

	return loadMinimalTaskView(tx, id, task)
}

func loadFullTaskView(tx *bolt.Tx, id string, task *tes.Task) error {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	if b == nil {
		return errNotFound
	}
	proto.Unmarshal(b, task)
	loadTaskLogs(tx, task)
	return loadMinimalTaskView(tx, id, task)
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
func (taskBolt *BoltDB) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task
	var err error

	err = taskBolt.db.View(func(tx *bolt.Tx) error {
		task, err = getTaskView(tx, req.Id, req.View)
		return err
	})

	if err != nil {
		log.Error("GetTask", "error", err, "taskID", req.Id)
		if err == errNotFound {
			return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", err.Error(), req.Id))
		}
	}
	return task, err
}

func getTaskView(tx *bolt.Tx, id string, view tes.TaskView) (*tes.Task, error) {
	var err error
	task := &tes.Task{}

	switch {
	case view == tes.TaskView_MINIMAL:
		err = loadMinimalTaskView(tx, id, task)
	case view == tes.TaskView_BASIC:
		err = loadBasicTaskView(tx, id, task)
	case view == tes.TaskView_FULL:
		err = loadFullTaskView(tx, id, task)
	default:
		err = fmt.Errorf("Unknown view: %s", view.String())
	}
	return task, err
}

// ListTasks returns a list of taskIDs
func (taskBolt *BoltDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

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
			task, _ := getTaskView(tx, string(k), req.View)
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
func (taskBolt *BoltDB) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	log := log.WithFields("taskID", req.Id)
	log.Info("Canceling task")

	// Check that the task exists
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		_, err := getTaskView(tx, req.Id, tes.TaskView_MINIMAL)
		return err
	})
	if err != nil {
		log.Error("CancelTask", "error", err, "taskID", req.Id)
		if err == errNotFound {
			return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", err.Error(), req.Id))
		}
	}

	err = taskBolt.db.Update(func(tx *bolt.Tx) error {
		// TODO need a test that ensures a canceled task is deleted from the worker
		return transitionTaskState(tx, req.Id, tes.State_CANCELED)
	})
	if err != nil {
		return nil, err
	}

	return &tes.CancelTaskResponse{}, nil
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
func (taskBolt *BoltDB) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{Name: taskBolt.conf.Server.ServiceName}, nil
}
