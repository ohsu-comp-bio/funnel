package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	sched_mocks "github.com/ohsu-comp-bio/funnel/scheduler/mocks"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"testing"
	"time"
)

func testRunnerFactoryFunc(f func(r testRunner)) RunnerFactory {
	return func(config.Worker, string) Runner {
		t := testRunner{}
		f(t)
		return &t
	}
}

type testRunner struct{}

func (t *testRunner) Run(context.Context) {}
func (t *testRunner) Factory(config.Worker, string) Runner {
	return t
}

// testWorker wraps Worker with some testing helpers.
type testWorker struct {
	*Worker
	Sched *sched_mocks.Client
	done  chan struct{}
}

func newTestWorker(conf config.Worker) testWorker {

	conf.WorkDir, _ = ioutil.TempDir("", "funnel-test-storage-")

	log.Debug("WORKDIR", conf.WorkDir)
	err := util.EnsureDir(conf.WorkDir)
	if err != nil {
		panic(err)
	}

	// A mock scheduler client allows this code to fake/control the worker's
	// communication with a scheduler service.
	s := new(sched_mocks.Client)
	w := &Worker{
		conf:      conf,
		sched:     s,
		log:       log,
		resources: detectResources(conf),
		newRunner: NoopRunnerFactory,
		runners:   runSet{},
		timeout:   util.NewIdleTimeout(conf.Timeout),
		stop:      make(chan struct{}),
		state:     pbf.WorkerState_ALIVE,
	}

	s.On("UpdateWorker", mock.Anything, mock.Anything).
		Return(nil, nil)
	s.On("UpdateWorker", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.On("Close").Return(nil)

	return testWorker{
		Worker: w,
		Sched:  s,
		done:   make(chan struct{}),
	}
}

func (t *testWorker) Start() {
	t.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.Worker{}, nil)
	go func() {
		t.Worker.Run()
		close(t.done)
	}()
}

func (t *testWorker) Wait() {
	<-t.done
}

func (t *testWorker) AddTasks(ids ...string) {
	// Set up the scheduler mock to assign a task to the worker.
	t.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.Worker{
			TaskIds: ids,
		}, nil).
		Once()

	t.Sched.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbf.Worker{}, nil)
}

func timeLimit(t *testing.T, d time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		select {
		case <-time.NewTimer(d).C:
			t.Fatal("time limit expired")
		case <-stop:
		}
	}()
	return func() {
		close(stop)
	}
}
