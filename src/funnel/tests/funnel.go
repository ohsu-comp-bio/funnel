package tests

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"funnel/scheduler"
	"funnel/scheduler/noop"
	"funnel/server"
	"funnel/server/mocks"
	"funnel/worker"
	"golang.org/x/net/context"
	"io/ioutil"
	"path"
)

// NewConfig returns the default config with a random port,
// and other common config used in tests.
func NewConfig() config.Config {
	conf := config.DefaultConfig()
	f, _ := ioutil.TempDir("", "funnel-test-")
	conf.WorkDir = f
	conf.DBPath = path.Join(f, "funnel.db")
	conf = servermocks.Config(conf)
	conf = noop.Config(conf)
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
	srv, err := server.NewServer(db, conf)
	if err != nil {
		panic(err)
	}

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
	m.Server.Start(ctx)
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
	m.Server.Stop()
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

// RunTask adds a task to the database (calling db.RunTask)
func (m *Funnel) RunTask(t *tes.Task) string {
	ret, err := m.DB.RunTask(context.Background(), t)
	if err != nil {
		panic(err)
	}
	return ret.Value
}

// RunHelloWorld adds a simple hello world task to the database queue.
func (m *Funnel) RunHelloWorld() string {
	return m.RunTask(m.HelloWorldTask())
}

// HelloWorldTask returns a simple hello world task.
func (m *Funnel) HelloWorldTask() *tes.Task {
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
func (m *Funnel) GetWorkers() []*pbf.Worker {
	resp, _ := m.DB.GetWorkers(context.Background(), &pbf.GetWorkersRequest{})
	return resp.Workers
}

// CompleteJob marks a job as completed
func (m *Funnel) CompleteJob(jobID string) {
	for _, w := range m.GetWorkers() {
		if j, ok := w.Jobs[jobID]; ok {
			j.Job.State = tes.State_Complete
			m.DB.UpdateWorker(context.Background(), w)
			return
		}
	}
	panic("No such job found: " + jobID)
}
