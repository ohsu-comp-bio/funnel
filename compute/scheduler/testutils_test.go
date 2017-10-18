package scheduler

import (
	"context"
	schedmock "github.com/ohsu-comp-bio/funnel/compute/scheduler/mocks"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"testing"
	"time"
)

func testWorkerFactoryFunc(f func(r testWorker)) WorkerFactory {
	return WorkerFactory(func(config.Worker, string, *logger.Logger) (worker.Worker, error) {
		t := testWorker{}
		f(t)
		return &t, nil
	})
}

type testWorker struct{}

func (t *testWorker) Run(context.Context) {}
func (t *testWorker) Factory(config.Worker, string, *logger.Logger) (worker.Worker, error) {
	return t, nil
}
func (t *testWorker) Close() error {
	return nil
}

// testNode wraps Node with some testing helpers.
type testNode struct {
	*Node
	Client *schedmock.Client
	done   chan struct{}
}

func newTestNode(conf config.Config, t *testing.T) testNode {
	workDir, _ := ioutil.TempDir("", "funnel-test-storage-")
	conf.Scheduler.Node.WorkDir = workDir
	conf.Worker.WorkDir = workDir
	log := logger.NewLogger("test-node", logger.DebugConfig())

	// A mock scheduler client allows this code to fake/control the worker's
	// communication with a scheduler service.
	res, _ := detectResources(conf.Scheduler.Node)
	s := new(schedmock.Client)
	n := &Node{
		conf:       conf.Scheduler.Node,
		workerConf: conf.Worker,
		client:     s,
		log:        log,
		resources:  res,
		newWorker:  NoopWorkerFactory,
		workers:    newRunSet(),
		timeout:    util.NewIdleTimeout(conf.Scheduler.Node.Timeout),
		state:      pbs.NodeState_ALIVE,
	}

	s.On("PutNode", mock.Anything, mock.Anything).
		Return(nil, nil)
	s.On("PutNode", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.On("Close").Return(nil)

	return testNode{
		Node:   n,
		Client: s,
		done:   make(chan struct{}),
	}
}

func (t *testNode) Start() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	t.Client.On("GetNode", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbs.Node{}, nil)
	go func() {
		t.Node.Run(ctx)
		close(t.done)
	}()
	return cancel
}

func (t *testNode) Wait() {
	<-t.done
}

func (t *testNode) AddTasks(ids ...string) {
	// Set up the scheduler mock to assign a task to the worker.
	t.Client.On("GetNode", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbs.Node{
			TaskIds: ids,
		}, nil).
		Once()

	t.Client.On("GetNode", mock.Anything, mock.Anything, mock.Anything).
		Return(&pbs.Node{}, nil)
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
