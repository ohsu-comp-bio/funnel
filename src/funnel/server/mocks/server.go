package mocks

import (
	"fmt"
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"funnel/scheduler"
	"funnel/server"
	"funnel/worker"
	"golang.org/x/net/context"
	"io/ioutil"
	"math/rand"
	"time"
)

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}

// NewConfig returns the default config with a random port,
// and other common config used in tests.
func NewConfig() config.Config {
	port := randomPort()
	conf := config.DefaultConfig()
	conf.RPCPort = port
	conf.Worker = config.WorkerInheritConfigVals(conf)
	// Write the database to a temporary file
	f, _ := ioutil.TempFile("", "funnel-test-db-")
	conf.DBPath = f.Name()
	conf.Scheduler = "noop"
	return conf
}

// NewServer creates a new test server, which includes helpers
// for the scheduler, noop backend, and lots of other utils.
func NewServer(conf config.Config) *Server {

	// Configuration
	conf.Worker = config.WorkerInheritConfigVals(conf)

	// Create database
	db, dberr := server.NewTaskBolt(conf)
	if dberr != nil {
		panic("Couldn't open database")
	}

	// Create server
	srv, err := server.NewServer(db, conf)
	if err != nil {
		panic(err)
	}

	sched, _ := scheduler.NewScheduler(db, conf)

	return &Server{
		DB:        db,
		Server:    srv,
		Scheduler: sched,
		Conf:      conf,
	}
}

// Server is a server to use during testing.
type Server struct {
	DB         *server.TaskBolt
	Server     *server.Server
	Scheduler  *scheduler.Scheduler
	Conf       config.Config
	NoopWorker *worker.Worker
	stop       context.CancelFunc
}

// Client returns a scheduler client.
func (m *Server) Client() scheduler.Client {
	// Create client
	client, err := scheduler.NewClient(m.Conf.Worker)
	if err != nil {
		panic(err)
	}
	return client
}

// Start starts the server and many subcomponents including
// the scheduler and noop backend.
func (m *Server) Start() {
	ctx, stop := context.WithCancel(context.Background())
	m.stop = stop
	m.Server.Start(ctx)
	m.NoopWorker = NewNoopWorker(m.Conf)
	m.Scheduler.AddBackend(scheduler.BackendFactory{
		Name: "noop",
		Create: func(conf config.Config) (scheduler.Backend, error) {
			return scheduler.Backend(&NoopBackend{m.NoopWorker, conf}), nil
		},
	})
}

// Stop stops the server and cleans up resources
func (m *Server) Stop() {
	if m.stop == nil {
		return
	}
	m.stop()
	m.stop = nil
	m.Client().Close()
	m.Server.Stop()
}

// Flush calls Schedule() and worker.Sync, which helps tests
// manually sync the server and worker instead of depending
// on tickers/timing.
func (m *Server) Flush() {
	s.Scheduler.Schedule(context.Background())
	s.NoopWorker.Sync()
}

// AddWorker adds the given worker to the database (calling db.UpdateWorker)
func (m *Server) AddWorker(w *pbf.Worker) {
	m.DB.UpdateWorker(context.Background(), w)
}

// RunTask adds a task to the database (calling db.RunTask)
func (m *Server) RunTask(t *tes.Task) string {
	ret, err := m.DB.RunTask(context.Background(), t)
	if err != nil {
		panic(err)
	}
	return ret.Value
}

// RunHelloWorld adds a simple hello world task to the database queue.
func (m *Server) RunHelloWorld() string {
	return m.RunTask(m.HelloWorldTask())
}

// HelloWorldTask returns a simple hello world task.
func (m *Server) HelloWorldTask() *tes.Task {
	return &tes.Task{
		Name: "Hello world",
		Docker: []*tes.DockerExecutor{
			{
				Cmd: []string{"echo", "hello world"},
			},
		},
		Resources: &tes.Resources{
			MinimumCpuCores: 1,
			Volumes: []*tes.Volume{
				{
					Name:       "test-vol",
					SizeGb:     10.0,
					MountPoint: "/tmp",
				},
			},
		},
	}
}

// GetWorkers calls db.GetWorkers.
func (m *Server) GetWorkers() []*pbf.Worker {
	resp, _ := m.DB.GetWorkers(context.Background(), &pbf.GetWorkersRequest{})
	return resp.Workers
}

// CompleteJob marks a job as completed
func (m *Server) CompleteJob(jobID string) {
	for _, w := range m.GetWorkers() {
		if j, ok := w.Jobs[jobID]; ok {
			j.Job.State = tes.State_Complete
			m.DB.UpdateWorker(context.Background(), w)
			return
		}
	}
	panic("No such job found: " + jobID)
}

func randomPort() string {
	min := 10000
	max := 20000
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}
