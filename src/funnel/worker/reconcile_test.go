package worker

import (
	"errors"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"testing"
)

func TestReconcileSingleTaskCompleteFlow(t *testing.T) {

	var err error
	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	err = w.reconcile(tasks)

	if err != nil {
		t.Error("Unexpected error on empty reconcile")
	}
	if len(w.Ctrls) != 0 {
		t.Error("Unexpected runner created on empty reconcile")
	}

	j := &tes.Task{
		TaskID: "task-1",
		State:  Queued,
	}
	addTask(tasks, j)

	w.reconcile(tasks)

	if _, exists := w.Ctrls["task-1"]; !exists {
		t.Error("Expected runner to be created for new task")
	}

	ctrl := w.Ctrls["task-1"]

	if j.State != Initializing {
		t.Error("Expected task just started to be in initializing state.")
	}

	ctrl.SetRunning()
	w.reconcile(tasks)

	if j.State != Running {
		t.Error("Expected task state to be running")
	}

	ctrl.SetResult(nil)
	w.reconcile(tasks)

	if j.State != Complete {
		t.Error("Expected task state to be complete")
	}
}

func TestReconcileTaskError(t *testing.T) {

	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}
	j := &tes.Task{
		TaskID: "task-1",
		State:  Queued,
	}
	addTask(tasks, j)
	w.reconcile(tasks)
	ctrl := w.Ctrls["task-1"]
	ctrl.SetResult(errors.New("Test task error"))
	w.reconcile(tasks)

	if j.State != Error {
		t.Error("Expected task state to be Error")
	}
}

func TestReconcileCancelTask(t *testing.T) {
	// Set up worker with no-op runner
	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	// Add a task
	j := &tes.Task{
		TaskID: "task-1",
		State:  Queued,
	}
	addTask(tasks, j)

	// Reconcile worker state, which registers task with worker
	w.reconcile(tasks)

	// Cancel task
	j.State = Canceled
	ctrl := w.Ctrls["task-1"]

	// Reconcile again. Worker should react to task being canceled.
	w.reconcile(tasks)

	if ctrl.State() != Canceled {
		t.Error("Expected runner state to be canceled")
	}

	// Delete canceled task. This is emulating what would happen with
	// the server. The worker won't delete a canceled task controller
	// until the server deletes the task first.
	delete(tasks, "task-1")
	w.reconcile(tasks)

	if w.Ctrls["task-1"] != nil {
		t.Error("Expected task ctrl to be cleaned up")
	}
}

func TestReconcileMultiple(t *testing.T) {

	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	w.reconcile(tasks)

	addTask(tasks, &tes.Task{
		TaskID: "task-1",
		State:  Queued,
	})

	w.reconcile(tasks)

	if _, exists := w.Ctrls["task-1"]; !exists {
		t.Error("Expected runner to be created for new task")
	}

	if tasks["task-1"].Task.State != Initializing {
		t.Error("Expected task just started to be in initializing state.")
	}

	w.Ctrls["task-1"].SetRunning()
	w.reconcile(tasks)

	if tasks["task-1"].Task.State != Running {
		t.Error("Expected task state to be running")
	}

	addTask(tasks, &tes.Task{
		TaskID: "task-2",
		State:  Queued,
	})
	addTask(tasks, &tes.Task{
		TaskID: "task-3",
		State:  Queued,
	})

	w.reconcile(tasks)

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
	w.reconcile(tasks)

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
	w.reconcile(tasks)

	if w.Ctrls["task-2"] != nil {
		t.Error("Expected task 2 ctrl to be cleaned up")
	}

	if tasks["task-3"].Task.State != Error {
		t.Error("Expected task 3 state to be error")
	}
}

// Tests how the worker handles the case where it finds a task without a controller
// and the task state is not Queued (normal case), but is Initializing or Running
func TestStraightToRunning(t *testing.T) {
	tasks := map[string]*pbf.TaskWrapper{}
	w := Worker{
		TaskRunner: NoopTaskRunner,
		Ctrls:      map[string]TaskControl{},
	}

	addTask(tasks, &tes.Task{
		TaskID: "task-1",
		State:  Initializing,
	})
	addTask(tasks, &tes.Task{
		TaskID: "task-2",
		State:  Running,
	})

	w.reconcile(tasks)

	if _, exists := w.Ctrls["task-1"]; !exists {
		t.Error("Expected runner to be created for new task 1")
	}
	if _, exists := w.Ctrls["task-2"]; !exists {
		t.Error("Expected runner to be created for new task 2")
	}

	if tasks["task-1"].Task.State != Initializing {
		t.Error("Expected task 1 state to be unchanged.")
	}

	if tasks["task-2"].Task.State != Initializing {
		t.Error("Expected task 2 state to revert to initializing.")
	}
}

// TODO test edge cases
// - missing task
// - missing ctrl
// - complete task, ctrl incomplete
