package boltdb

import (
	"github.com/boltdb/bolt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
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

// ExecutorStdout maps (task ID + executor index) -> tes.ExecutorLog.Stdout string
var ExecutorStdout = []byte("executor-stdout")

// ExecutorStderr maps (task ID + executor index) -> tes.ExecutorLog.Stderr string
var ExecutorStderr = []byte("executor-stderr")

// Nodes maps:
// node ID -> pbs.Node struct
var Nodes = []byte("nodes")

// SysLogs defeines the name of a bucket with maps
//  task ID -> tes.TaskLog.SystemLogs
var SysLogs = []byte("system-logs")

// BoltDB provides handlers for gRPC endpoints.
// Data is stored/retrieved from the BoltDB key-value database.
type BoltDB struct {
	db   *bolt.DB
	conf config.Config
}

// NewBoltDB returns a new instance of BoltDB, accessing the database at
// the given path, and including the given ServerConfig.
func NewBoltDB(conf config.Config) (*BoltDB, error) {
	fsutil.EnsurePath(conf.Server.Databases.BoltDB.Path)
	db, err := bolt.Open(conf.Server.Databases.BoltDB.Path, 0600, &bolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}

	b := &BoltDB{db: db, conf: conf}
	if err := b.init(); err != nil {
		return nil, err
	}
	return b, nil
}

// init creates the required BoltDB buckets
func (taskBolt *BoltDB) init() error {
	// Check to make sure all the required buckets have been created
	return taskBolt.db.Update(func(tx *bolt.Tx) error {
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
		if tx.Bucket(ExecutorStdout) == nil {
			tx.CreateBucket(ExecutorStdout)
		}
		if tx.Bucket(ExecutorStderr) == nil {
			tx.CreateBucket(ExecutorStderr)
		}
		if tx.Bucket(Nodes) == nil {
			tx.CreateBucket(Nodes)
		}
		if tx.Bucket(SysLogs) == nil {
			tx.CreateBucket(SysLogs)
		}
		return nil
	})
}
