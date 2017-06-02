package worker

import (
	"errors"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"testing"
)

func TestSyncSingleTaskCompleteFlow(t *testing.T) {

	var err error
	tasks := map[string]*tes.Task{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
	}

	w.Sync()

	if w.runners.Count() != 0 {
		t.Error("Unexpected runner created on empty reconcile")
	}

	j := &tes.Task{
		Id:    "task-1",
		State: Queued,
	}
	addTask(tasks, j)

	w.Sync()

	if w.runners.Count() != 1 {
		t.Error("Expected runner to be created for new task")
	}

	ctrl := w.Ctrls["task-1"]

	if j.State != Initializing {
		t.Error("Expected task just started to be in initializing state.")
	}

	ctrl.SetRunning()
	w.Sync()

	if j.State != Running {
		t.Error("Expected task state to be running")
	}

	ctrl.SetResult(nil)

	if j.State != Complete {
		t.Error("Expected task state to be complete")
	}
}

func TestSyncTaskError(t *testing.T) {

	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}
	j := &tes.Task{
		Id:    "task-1",
		State: Queued,
	}
	addTask(tasks, j)
	w.Sync()
	ctrl := w.Ctrls["task-1"]
	ctrl.SetResult(errors.New("Test task error"))
	w.Sync()

	if j.State != Error {
		t.Error("Expected task state to be Error")
	}
}

func TestSyncCancelTask(t *testing.T) {
	// Set up worker with no-op runner
	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	// Add a task
	j := &tes.Task{
		Id:    "task-1",
		State: Queued,
	}
	addTask(tasks, j)

	w.Sync()

	// Cancel task
	j.State = Canceled
	ctrl := w.Ctrls["task-1"]

	w.Sync()

	if ctrl.State() != Canceled {
		t.Error("Expected runner state to be canceled")
	}

	// Delete canceled task. This is emulating what would happen with
	// the server. The worker won't delete a canceled task controller
	// until the server deletes the task first.
	delete(tasks, "task-1")
	w.Sync()

	if w.Ctrls["task-1"] != nil {
		t.Error("Expected task ctrl to be cleaned up")
	}
}

func TestSyncMultiple(t *testing.T) {

	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	w.Sync()

	addTask(tasks, &tes.Task{
		Id:    "task-1",
		State: Queued,
	})

	w.Sync()

	if _, exists := w.Ctrls["task-1"]; !exists {
		t.Error("Expected runner to be created for new task")
	}

	if tasks["task-1"].Task.State != Initializing {
		t.Error("Expected task just started to be in initializing state.")
	}

	w.Ctrls["task-1"].SetRunning()
	w.Sync()

	if tasks["task-1"].Task.State != Running {
		t.Error("Expected task state to be running")
	}

	addTask(tasks, &tes.Task{
		Id:    "task-2",
		State: Queued,
	})
	addTask(tasks, &tes.Task{
		Id:    "task-3",
		State: Queued,
	})

	w.Sync()

	if len(w.Ctrls) != 3 {
		t.Error("Expected runner to be created for new task")
	}

	if tasks["task-2"].Task.State != Initializing {
		t.Error("Expected task 2 state to be init")
	}

	if tasks["task-3"].Task.State != Initializing {
		t.Error("Expected task 3 state to be init")
	}

	// Set task-1 to complete
	w.Ctrls["task-1"].SetResult(nil)
	// Cancel task-2
	tasks["task-2"].Task.State = Canceled
	// Set task-3 to error
	w.Ctrls["task-3"].SetResult(errors.New("Task 3 error"))

	j2ctrl := w.Ctrls["task-2"]
	w.Sync()

	if tasks["task-1"].Task.State != Complete {
		t.Error("Expected task 1 state to be complete")
	}

	if j2ctrl.State() != Canceled {
		t.Error("Expected task 2 controller to be canceled state")
	}

	// Delete canceled task. This is emulating what would happen with
	// the server. The worker won't delete a canceled task controller
	// until the server deletes the task first.
	delete(tasks, "task-2")
	w.Sync()

	if w.Ctrls["task-2"] != nil {
		t.Error("Expected task 2 ctrl to be cleaned up")
	}

	if tasks["task-3"].Task.State != Error {
		t.Error("Expected task 3 state to be error")
	}
}
