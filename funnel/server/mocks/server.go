package mocks

import (
	"fmt"
	"funnel/config"
	pbf "funnel/proto/funnel"
	tes "funnel/proto/tes"
	"funnel/scheduler"
	"funnel/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"math/rand"
	"net"
	"time"
)

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}

// NewMockServerConfig returns the default config with a random port
func NewMockServerConfig() config.Config {
	port := randomPort()
	conf := config.DefaultConfig()
	conf.RPCPort = port
	conf.Worker = config.WorkerInheritConfigVals(conf)
	return conf
}

// NewMockServer starts a test server. This creates a database in a temp. file
// and starts a gRPC server on a random port.
func NewMockServer() *MockServer {
	conf := NewMockServerConfig()
	return MockServerFromConfig(conf)
}

// MockServerFromConfig starts a test server with the given config.
func MockServerFromConfig(conf config.Config) *MockServer {
	// Write the database to a temporary file
	f, _ := ioutil.TempFile("", "funnel-test-db-")

	// Configuration
	conf.Worker = config.WorkerInheritConfigVals(conf)
	conf.DBPath = f.Name()

	// Create database
	db, dberr := server.NewTaskBolt(conf)
	if dberr != nil {
		panic("Couldn't open database")
	}

	// Listen on TCP port for RPC
	server := grpc.NewServer()
	lis, err := net.Listen("tcp", ":"+conf.RPCPort)
	if err != nil {
		panic("Cannot open port: " + conf.RPCPort)
	}

	// Create client
	client, err := scheduler.NewClient(conf.Worker)
	if err != nil {
		panic("Can't connect scheduler client")
	}

	pbf.RegisterSchedulerServer(server, db)
	go server.Serve(lis)

	return &MockServer{
		DB:     db,
		Client: client,
		srv:    server,
		Conf:   conf,
	}
}

// MockServer is a server to use during testing.
type MockServer struct {
	DB     *server.TaskBolt
	Client scheduler.Client
	srv    *grpc.Server
	Conf   config.Config
}

// Close cleans up the mock server resources
func (m *MockServer) Close() {
	m.Client.Close()
	m.srv.GracefulStop()
}

// AddWorker adds the given worker to the database (calling db.UpdateWorker)
func (m *MockServer) AddWorker(w *pbf.Worker) {
	m.DB.UpdateWorker(context.Background(), w)
}

// RunTask adds a task to the database (calling db.RunTask)
func (m *MockServer) RunTask(t *tes.Task) string {
	ret, err := m.DB.RunTask(context.Background(), t)
	if err != nil {
		panic(err)
	}
	return ret.Value
}

// RunHelloWorld adds a simple hello world task to the database queue.
func (m *MockServer) RunHelloWorld() string {
	return m.RunTask(m.HelloWorldTask())
}

// HelloWorldTask returns a simple hello world task.
func (m *MockServer) HelloWorldTask() *tes.Task {
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
func (m *MockServer) GetWorkers() []*pbf.Worker {
	resp, _ := m.DB.GetWorkers(context.Background(), &pbf.GetWorkersRequest{})
	return resp.Workers
}

// CompleteJob marks a job as completed
func (m *MockServer) CompleteJob(jobID string) {
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
