package worker

import (
	"errors"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func init() {
	log.Configure(logger.DebugConfig())
}

// Test calling Worker.Stop()
func TestStopWorker(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	w.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.Worker{}, nil)

	w.Start()

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, time.Millisecond*4)
	defer cleanup()
	w.Stop()
	w.Wait()
	w.Sched.AssertCalled(t, "Close")
}

// Mainly exercising a panic bug caused by an unhandled
// error from client.GetWorker().
func TestGetWorkerFail(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	// Set GetWorker to return an error
	w.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("TEST"))
	w.sync()
	time.Sleep(time.Second)
}

// Test the flow of a worker completing a task then timing out
func TestWorkerTimeout(t *testing.T) {
	conf := config.DefaultConfig().Worker
	conf.Timeout = time.Millisecond
	conf.UpdateRate = time.Millisecond * 2

	w := newTestWorker(conf)

	// Set up a test runner which this code can easily control.
	r := testRunner{}
	// Hook the test runner up to the worker's runner factory.
	w.newRunner = r.Factory

	// Set up scheduler mock to return a task
	w.AddTasks("task-1")

	w.Start()

	// Fail if this test doesn't complete in the given time.
	cleanup := timeLimit(t, conf.Timeout*100)
	defer cleanup()

	// Wait for the worker to exit
	w.Wait()
}

// Test that a worker does nothing where there are no assigned tasks.
func TestNoTasks(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	// Tell the scheduler mock to return nothing
	w.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.Worker{}, nil)

	// Count the number of times the runner factory was called
	var count int
	// Hook the test runner up to the worker's runner factory.
	w.newRunner = testRunnerFactoryFunc(func(r testRunner) {
		count++
	})

	w.sync()
	w.sync()
	w.sync()
	time.Sleep(time.Second)

	if count != 0 {
		t.Fatal("Unexpected runner factory call count")
	}
	if w.runners.Count() != 0 {
		t.Fatal("Unexpected worker runner count")
	}
}

// Test that a runner gets created for each task.
func TestWorkerRunnerCreated(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	// Count the number of times the runner factory was called
	var count int
	// Hook the test runner up to the worker's runner factory.
	w.newRunner = testRunnerFactoryFunc(func(r testRunner) {
		count++
	})

	w.AddTasks("task-1", "task-2")
	w.sync()
	time.Sleep(time.Second)

	log.Debug("COUNT", count)
	if count != 2 {
		t.Fatal("Unexpected worker runner count")
	}
}

// Test that a finished task is not immediately re-run.
// Tests a bugfix.
func TestFinishedTaskNotRerun(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	// Set up a test runner which this code can easily control.
	r := testRunner{}
	// Hook the test runner up to the worker's runner factory.
	w.newRunner = r.Factory

	w.AddTasks("task-1")

	// manually sync the worker to avoid timing issues.
	w.sync()
	time.Sleep(time.Second)

	log.Debug("COUNT", w.runners.Count())
	if w.runners.Count() != 0 {
		t.Fatal("Unexpected runner count")
	}

	// There was a bug where later syncs would end up re-running the task.
	// Do a few syncs to make sure.
	w.sync()
	w.sync()
	time.Sleep(time.Second)

	log.Debug("COUNT", w.runners.Count())
	if w.runners.Count() != 0 {
		t.Fatal("Unexpected runner count")
	}
}

// Test that tasks are removed from the worker's runset when they finish.
func TestFinishedTaskRunsetCount(t *testing.T) {
	conf := config.DefaultConfig().Worker
	w := newTestWorker(conf)

	// Set up a test runner which this code can easily control.
	r := testRunner{}
	// Hook the test runner up to the worker's runner factory.
	w.newRunner = r.Factory

	w.AddTasks("task-1")

	// manually sync the worker to avoid timing issues.
	w.sync()
	time.Sleep(time.Second)

	if w.runners.Count() != 0 {
		log.Debug("COUNT", w.runners.Count())
		t.Fatal("Unexpected runner count")
	}
}
