package boltdb

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
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

// ExecutorStdout maps (task ID + executor index) -> tes.ExecutorLog.Stdout string
var ExecutorStdout = []byte("executor-stdout")

// ExecutorStderr maps (task ID + executor index) -> tes.ExecutorLog.Stderr string
var ExecutorStderr = []byte("executor-stderr")

// Nodes maps:
// node ID -> pbs.Node struct
var Nodes = []byte("nodes")

// SysLogs defeines the name of a bucket with maps
//
//	task ID -> tes.TaskLog.SystemLogs
var SysLogs = []byte("system-logs")

// BoltDB provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type BoltDB struct {
	scheduler.UnimplementedSchedulerServiceServer
	db *bolt.DB
}

// NewBoltDB returns a new instance of BoltDB, accessing the database at
// the given path, and including the given ServerConfig.
func NewBoltDB(conf config.BoltDB) (*BoltDB, error) {
	err := fsutil.EnsurePath(conf.Path)
	if err != nil {
		return nil, err
	}
	fmt.Println("DEBUG conf:", conf)
	db, err := bolt.Open(conf.Path, 0600, &bolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}
	return &BoltDB{db: db}, nil
}

// Init creates the required BoltDB buckets
func (taskBolt *BoltDB) Init() error {
	// Check to make sure all the required buckets have been created
	return taskBolt.db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(TaskBucket) == nil {
			_, err := tx.CreateBucket(TaskBucket)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(TasksQueued) == nil {
			_, err := tx.CreateBucket(TasksQueued)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(TaskState) == nil {
			_, err := tx.CreateBucket(TaskState)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(TasksLog) == nil {
			_, err := tx.CreateBucket(TasksLog)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(ExecutorLogs) == nil {
			_, err := tx.CreateBucket(ExecutorLogs)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(ExecutorStdout) == nil {
			_, err := tx.CreateBucket(ExecutorStdout)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(ExecutorStderr) == nil {
			_, err := tx.CreateBucket(ExecutorStderr)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(Nodes) == nil {
			_, err := tx.CreateBucket(Nodes)
			if err != nil {
				return err
			}
		}
		if tx.Bucket(SysLogs) == nil {
			_, err := tx.CreateBucket(SysLogs)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (taskBolt *BoltDB) Close() {
	taskBolt.db.Close()
}
