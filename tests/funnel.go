package tests

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/scheduler/noop"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/worker"
	"golang.org/x/net/context"
	"io/ioutil"
	"math/rand"
	"path"
	"time"
)

// NewConfig returns the default config with a random port,
// and other common config used in tests.
func NewConfig() config.Config {
	conf := config.DefaultConfig()
	f, _ := ioutil.TempDir("", "funnel-test-")
	conf.WorkDir = f
	conf.DBPath = path.Join(f, "funnel.db")
	conf = noop.Config(conf)
	conf.RPCPort = randomPort()
	conf.HTTPPort = randomPort()
	conf.Worker = config.WorkerInheritConfigVals(conf)
	return conf
}

// NewFunnel creates a new test server, which includes helpers
// for the scheduler, noop backend, and lots of other utils.
func NewFunnel(conf config.Config) *Funnel {

	// Configuration
	conf.Worker = config.WorkerInheritConfigVals(conf)

	// Create database
	db, dberr := server.NewTaskBolt(conf)
	if dberr != nil {
		panic("Couldn't open database")
	}

	// Create server
	srv := server.DefaultServer(db, conf)
	sched, _ := scheduler.NewScheduler(db, conf)

	return &Funnel{
		DB:        db,
		Server:    srv,
		Scheduler: sched,
		Conf:      conf,
	}
}

// Funnel is a server to use during testing.
type Funnel struct {
	DB         *server.TaskBolt
	Server     *server.Server
	Scheduler  *scheduler.Scheduler
	Conf       config.Config
	NoopWorker *worker.Worker
	stop       context.CancelFunc
}

// Client returns a scheduler client.
func (m *Funnel) Client() scheduler.Client {
	// Create client
	client, err := scheduler.NewClient(m.Conf.Worker)
	if err != nil {
		panic(err)
	}
	return client
}

// Start starts the server and many subcomponents including
// the scheduler and noop backend.
func (m *Funnel) Start() {
	ctx, stop := context.WithCancel(context.Background())
	m.stop = stop
	go m.Server.Serve(ctx)
	time.Sleep(time.Millisecond * 300)
	m.NoopWorker = noop.NewWorker(m.Conf)
	m.Scheduler.AddBackend(noop.NewPlugin(m.NoopWorker))
}

// Stop stops the server and cleans up resources
func (m *Funnel) Stop() {
	if m.stop == nil {
		return
	}
	m.stop()
	m.stop = nil
	m.Client().Close()
}

// Flush calls Schedule() and worker.Sync, which helps tests
// manually sync the server and worker instead of depending
// on tickers/timing.
func (m *Funnel) Flush() {
	m.Scheduler.Schedule(context.Background())
	m.NoopWorker.Sync()
}

// AddWorker adds the given worker to the database (calling db.UpdateWorker)
func (m *Funnel) AddWorker(w *pbf.Worker) {
	m.DB.UpdateWorker(context.Background(), w)
}

// CreateTask adds a task to the database (calling db.CreateTask)
func (m *Funnel) CreateTask(t *tes.Task) string {
	ret, err := m.DB.CreateTask(context.Background(), t)
	if err != nil {
		panic(err)
	}
	return ret.Id
}

// RunHelloWorld adds a simple hello world task to the database queue.
func (m *Funnel) RunHelloWorld() string {
	return m.CreateTask(m.HelloWorldTask())
}

// HelloWorldTask returns a simple hello world task.
func (m *Funnel) HelloWorldTask() *tes.Task {
	return &tes.Task{
		Name: "Hello world",
		Executors: []*tes.Executor{
			{
				Cmd: []string{"echo", "hello world"},
			},
		},
		Resources: &tes.Resources{
			CpuCores: 1,
		},
	}
}

// ListWorkers calls db.ListWorkers.
func (m *Funnel) ListWorkers() []*pbf.Worker {
	resp, _ := m.DB.ListWorkers(context.Background(), &pbf.ListWorkersRequest{})
	return resp.Workers
}

// CompleteTask marks a task as completed
func (m *Funnel) CompleteTask(taskID string) {
	for _, w := range m.ListWorkers() {
		if j, ok := w.Tasks[taskID]; ok {
			j.Task.State = tes.State_COMPLETE
			m.DB.UpdateWorker(context.Background(), w)
			return
		}
	}
	panic("No such task found: " + taskID)
}

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}
func randomPort() string {
	min := 10000
	max := 20000
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}
