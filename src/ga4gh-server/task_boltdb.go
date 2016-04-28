
package ga4gh_task

import (
	"golang.org/x/net/context"
	"ga4gh-tasks"
	"github.com/boltdb/bolt"
	uuid "github.com/nu7hatch/gouuid"
	proto "github.com/golang/protobuf/proto"
)


var TASK_BUCKET = []byte("tasks")
var TASKOP_BUCKET = []byte("taskops")
var TASKOP_MAP_BUCKET = []byte("taskop-map")
var TASKOP_STATUS_BUCKET = []byte("taskop-status")

type TaskBolt struct {
	db *bolt.DB
}


func NewTaskBolt(path string) *TaskBolt {
	db, _ := bolt.Open(path, 0600, nil)
	return &TaskBolt{db:db}
}


func (self *TaskBolt) CreateTask(ctx context.Context, task *ga4gh_task_exec.Task) (*ga4gh_task_exec.Task, error) {
	u, _ := uuid.NewV4()
	task.TaskId = u.String()
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		err := b.Put(task.TaskId, proto.Marshal(task) )
		return err
	})
	return task, err
}

// / Delete a task
func (self *TaskBolt) DeleteTask(ctx context.Context, task *ga4gh_task_exec.TaskId) (*ga4gh_task_exec.TaskId, error) {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		err := b.Delete([]byte(task.Value))
		return err
	})
	return task, err
}

// / Get a task by its ID
func (self *TaskBolt) GetTask(ctx context.Context, task *ga4gh_task_exec.TaskId) (*ga4gh_task_exec.TaskId, error) {
	ch := make(chan *ga4gh_task_exec.Task)
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		v := b.Get([]byte(task.Value))
		out := ga4gh_task_exec.Task{}
		proto.Unmarshal(v, &out)
		ch <- out
		return nil
	})
	a := <- ch
	return a, nil
}

// / Run a task
func (self *TaskBolt) RunTask(ctx context.Context, request *ga4gh_task_exec.TaskRunRequest) (*ga4gh_task_exec.TaskOpId, error) {
	ch := make(chan *ga4gh_task_exec.TaskOpId)
	self.db.Update(func(tx *bolt.Tx) error {
		taskopId, _ := uuid.NewV4()
		taskop := ga4gh_task_exec.TaskOp{
			Name: taskopId.String(),
			Done:false,
		}

		taskop_b := tx.Bucket(TASKOP_BUCKET)
		taskop_b.Put( []byte(taskopId.String()), proto.Marshal(taskop) )

		ch <- &ga4gh_task_exec.TaskOpId{Value:taskopId.String()}
		return nil
	})
	a := <- ch
	return a, nil
}

// / Get info about a running task
func (self *TaskBolt) GetTaskOp(ctx context.Context, taskop *ga4gh_task_exec.TaskOpId) (*ga4gh_task_exec.TaskOp, error) {
	ch := make(chan *ga4gh_task_exec.TaskOp)
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASKOP_BUCKET)
		v := b.Get([]byte(taskop.Value))
		out := ga4gh_task_exec.TaskOp{}
		proto.Unmarshal(v, &out)
		ch <- out
		return nil
	})
	a := <- ch
	return a, nil
}

// / Cancel a running task
func (self *TaskBolt) CancelTaskOp(ctx context.Context, taskop *ga4gh_task_exec.TaskOpId) (*ga4gh_task_exec.TaskOpId, error) {
	self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASKOP_STATUS_BUCKET)
		b.Put([]byte(taskop.Value), []byte("Canceled"))
		return nil
	})
	return taskop
}