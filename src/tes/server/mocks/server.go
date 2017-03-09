package mocks

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"math/rand"
	"net"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/scheduler"
	"tes/server"
	pbr "tes/server/proto"
	"time"
)

func NewMockServer() *MockServer {
	// Write the database to a temporary file
	f, _ := ioutil.TempFile("", "funnel-test-db-")

	// Configuration
	port := randomPort()
	conf := config.DefaultConfig()
	conf.ServerAddress = "localhost:" + port
	conf.Worker.ServerAddress = conf.ServerAddress
	conf.RPCPort = port
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
		panic("Cannot open port")
	}

	// Create client
	client, err := scheduler.NewClient(conf.Worker)
	if err != nil {
		panic("Can't connect scheduler client")
	}

	pbr.RegisterSchedulerServer(server, db)
	go server.Serve(lis)

	return &MockServer{
		DB:     db,
		Client: client,
    srv: server,
	}
}

type MockServer struct {
	DB     *server.TaskBolt
	Client scheduler.Client
  srv    *grpc.Server
}

func (m *MockServer) Close() {
  m.srv.Stop()
}

func (m *MockServer) AddWorker(w *pbr.Worker) {
	m.DB.UpdateWorker(context.Background(), w)
}

func (m *MockServer) RunTask(t *pbe.Task) {
	_, err := m.DB.RunTask(context.Background(), t)
	if err != nil {
		panic(err)
	}
}

func (m *MockServer) RunHelloWorld() {
	m.RunTask(&pbe.Task{
		Name: "Hello world",
		Docker: []*pbe.DockerExecutor{
			{
				Cmd: []string{"echo", "hello world"},
			},
		},
		Resources: &pbe.Resources{
			MinimumCpuCores: 1,
		},
	})
}

func (m *MockServer) GetWorkers() []*pbr.Worker {
	resp, _ := m.DB.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})
	return resp.Workers
}

func randomPort() string {
	min := 10000
	max := 11000
	rand.Seed(time.Now().Unix())
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}
