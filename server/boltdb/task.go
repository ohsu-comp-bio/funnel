package boltdb

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

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
		return tes.ErrNotFound
	}
	task.Id = id
	task.State = getTaskState(tx, id)
	return nil
}

func loadBasicTaskView(tx *bolt.Tx, id string, task *tes.Task) error {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	if b == nil {
		return tes.ErrNotFound
	}
	proto.Unmarshal(b, task)
	loadTaskLogs(tx, task)

	// remove content from inputs
	inputs := []*tes.Input{}
	for _, v := range task.Inputs {
		v.Content = ""
		inputs = append(inputs, v)
	}
	task.Inputs = inputs

	return loadMinimalTaskView(tx, id, task)
}

func loadFullTaskView(tx *bolt.Tx, id string, task *tes.Task) error {
	b := tx.Bucket(TaskBucket).Get([]byte(id))
	if b == nil {
		return tes.ErrNotFound
	}
	proto.Unmarshal(b, task)
	loadTaskLogs(tx, task)

	// Load executor stdout/err
	for _, tl := range task.Logs {
		for j, el := range tl.Logs {
			key := fmt.Sprint(id, j)

			b := tx.Bucket(ExecutorStdout).Get([]byte(key))
			if b != nil {
				el.Stdout = string(b)
			}

			b = tx.Bucket(ExecutorStderr).Get([]byte(key))
			if b != nil {
				el.Stderr = string(b)
			}
		}
	}

	// Load system logs
	var syslogs []string
	slb := tx.Bucket(SysLogs).Get([]byte(id))
	if slb != nil {
		err := json.Unmarshal(slb, &syslogs)
		if err != nil {
			return err
		}
		task.Logs[0].SystemLogs = syslogs
	}

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
	pageSize := tes.GetPageSize(req.GetPageSize())

	taskBolt.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(TaskBucket).Cursor()

		i := 0

		// For pagination, figure out the starting key.
		var k []byte
		if req.PageToken != "" {
			// Seek moves to the key, but the start of the page is the next key.
			c.Seek([]byte(req.PageToken))
			k, _ = c.Prev()
		} else {
			// No pagination, so take the last key.
			// Keys (task IDs) are in ascending order, and we want the first page
			// to be the most recent task, so that's at the end of the list.
			k, _ = c.Last()
		}

	taskLoop:
		for ; k != nil && i < pageSize; k, _ = c.Prev() {
			task, _ := getTaskView(tx, string(k), req.View)
			if req.State != tes.Unknown && req.State != task.State {
				continue taskLoop
			}

			for k, v := range req.Tags {
				tval, ok := task.Tags[k]
				if !ok || tval != v {
					continue taskLoop
				}
			}

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
