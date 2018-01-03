package tes

import (
	"errors"
	"fmt"
	"github.com/getlantern/deepcopy"
	"github.com/golang/protobuf/jsonpb"
	"github.com/rs/xid"
	"time"
)

// Marshaler marshals tasks to indented JSON.
var Marshaler = jsonpb.Marshaler{
	Indent: "  ",
}

// MarshalToString marshals a task to an indented JSON string.
func MarshalToString(t *Task) (string, error) {
	if t == nil {
		return "", fmt.Errorf("can't marshal nil task")
	}
	return Marshaler.MarshalToString(t)
}

// ErrNotFound is returned when a task is not found.
var ErrNotFound = errors.New("task not found")

// Shorthand for task views
const (
	Minimal = TaskView_MINIMAL
	Basic   = TaskView_BASIC
	Full    = TaskView_FULL
)

// GenerateID generates a task ID string.
// IDs are globally unique and sortable.
func GenerateID() string {
	id := xid.New()
	return id.String()
}

// InitTask intializes task fields which are commonly set by CreateTask,
// such as Id, CreationTime, State, etc. If the task fails validation,
// an error is returned. See Validate().
// The given task is modified.
func InitTask(task *Task) error {
	task.Id = GenerateID()
	task.State = Queued
	task.CreationTime = time.Now().Format(time.RFC3339Nano)
	if err := Validate(task); err != nil {
		return fmt.Errorf("invalid task message:\n%s", err)
	}
	return nil
}

// RunnableState returns true if the state is RUNNING or INITIALIZING
func RunnableState(s State) bool {
	return s == State_INITIALIZING || s == State_RUNNING
}

// TerminalState returns true if the state is COMPLETE, ERROR, SYSTEM_ERROR, or CANCELED
func TerminalState(s State) bool {
	return s == State_COMPLETE || s == State_EXECUTOR_ERROR || s == State_SYSTEM_ERROR ||
		s == State_CANCELED
}

// GetBasicView returns the basic view of a task.
func (task *Task) GetBasicView() *Task {
	view := &Task{}
	deepcopy.Copy(view, task)

	// remove contents from inputs
	for _, v := range view.Inputs {
		v.Content = ""
	}

	// remove stdout and stderr from Task.Logs.Logs
	for _, tl := range view.Logs {
		for _, el := range tl.Logs {
			el.Stdout = ""
			el.Stderr = ""
		}
	}
	return view
}

// GetMinimalView returns the minimal view of a task.
func (task *Task) GetMinimalView() *Task {
	id := task.Id
	state := task.State
	return &Task{
		Id:    id,
		State: state,
	}
}

// GetTaskLog gets the task log entry at the given index "i".
// If the entry doesn't exist, empty logs will be appended up to "i".
func (task *Task) GetTaskLog(i int) *TaskLog {

	// Grow slice length if necessary
	for j := len(task.Logs); j <= i; j++ {
		task.Logs = append(task.Logs, &TaskLog{})
	}

	return task.Logs[i]
}

// GetExecLog gets the executor log entry at the given index "i".
// If the entry doesn't exist, empty logs will be appended up to "i".
func (task *Task) GetExecLog(attempt int, i int) *ExecutorLog {
	tl := task.GetTaskLog(attempt)

	// Grow slice length if necessary
	for j := len(tl.Logs); j <= i; j++ {
		tl.Logs = append(tl.Logs, &ExecutorLog{})
	}

	return tl.Logs[i]
}

// GetPageSize takes in the page size from a request and returns a new page size
// taking into account the minimum, maximum and default as documented in the TES spec.
func GetPageSize(reqSize uint32) int {
	// default page size
	var pageSize = 256

	if reqSize != 0 {
		pageSize = int(reqSize)

		// max page size
		if pageSize > 2048 {
			pageSize = 2048
		}

		// min page size
		if pageSize < 50 {
			pageSize = 50
		}
	}

	return pageSize
}
