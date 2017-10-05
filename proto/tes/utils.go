package tes

import (
	"github.com/getlantern/deepcopy"
)

// RunnableState returns true if the state is RUNNING or INITIALIZING
func RunnableState(s State) bool {
	return s == State_INITIALIZING || s == State_RUNNING
}

// TerminalState returns true if the state is COMPLETE, ERROR, SYSTEM_ERROR, or CANCELED
func TerminalState(s State) bool {
	return s == State_COMPLETE || s == State_ERROR || s == State_SYSTEM_ERROR ||
		s == State_CANCELED
}

// GetBasicView returns the basic view of a task.
func (task *Task) GetBasicView() *Task {
	view := &Task{}
	deepcopy.Copy(view, task)

	// remove contents from inputs
	for _, v := range view.Inputs {
		v.Contents = ""
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
