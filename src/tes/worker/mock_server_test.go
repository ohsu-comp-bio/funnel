package worker

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/scheduler"
	"tes/server"
	pbr "tes/server/proto"
)

type mockScheduler struct {
	worker *pbr.Worker
}

func (m *mockScheduler) Schedule(j *pbe.Job) *scheduler.Offer {
	return scheduler.NewOffer(m.worker, j, scheduler.Scores{})
}

func newMockSchedulerServer() *mockSchedulerServer {
	f, _ := ioutil.TempFile("", "funnel-test-db-")
	conf := config.DefaultConfig()
	conf.ServerAddress = "localhost:9932"
	conf.RPCPort = "9932"
	conf.DBPath = f.Name()

	db, dberr := server.NewTaskBolt(conf)
	if dberr != nil {
		panic("Couldn't open database")
	}

	server := grpc.NewServer()
	lis, err := net.Listen("tcp", ":"+conf.RPCPort)
	if err != nil {
		panic("Cannot open port")
	}

	wconf := config.WorkerDefaultConfig()
	wconf.ServerAddress = "localhost:9932"
	wconf.ID = "test-worker"
	x, werr := NewWorker(wconf)
	if werr != nil {
		panic(werr)
	}
	w := x.(*worker)
	w.runJob = noopRunJob

	sched := &mockScheduler{&pbr.Worker{
		Id:    "test-worker",
		State: pbr.WorkerState_Alive,
		Jobs:  map[string]*pbr.JobWrapper{},
	}}

	m := mockSchedulerServer{db, server, sched, conf, w}
	pbr.RegisterSchedulerServer(server, m)
	go server.Serve(lis)

	return &m
}

type mockSchedulerServer struct {
	db     *server.TaskBolt
	server *grpc.Server
	sched  *mockScheduler
	conf   config.Config
	worker *worker
}

func (m *mockSchedulerServer) Flush() {
	scheduler.Schedule(m.db, m.sched, m.conf)
	m.worker.checkJobs()
}

func (m *mockSchedulerServer) Stop() {
	m.server.Stop()
}

func (m mockSchedulerServer) UpdateWorker(ctx context.Context, req *pbr.Worker) (*pbr.UpdateWorkerResponse, error) {
	log.Debug("UpdateWorker", req)
	return m.db.UpdateWorker(ctx, req)
}
func (m mockSchedulerServer) GetWorker(ctx context.Context, req *pbr.GetWorkerRequest) (*pbr.Worker, error) {
	resp, err := m.db.GetWorker(ctx, req)
	log.Debug("GetWorker", "resp", resp)
	return resp, err
}
func (m mockSchedulerServer) GetWorkers(ctx context.Context, req *pbr.GetWorkersRequest) (*pbr.GetWorkersResponse, error) {
	return m.db.GetWorkers(ctx, req)
}
func (m mockSchedulerServer) UpdateJobLogs(ctx context.Context, req *pbr.UpdateJobLogsRequest) (*pbr.UpdateJobLogsResponse, error) {
	return m.db.UpdateJobLogs(ctx, req)
}
func (m mockSchedulerServer) GetQueueInfo(request *pbr.QueuedTaskInfoRequest, server pbr.Scheduler_GetQueueInfoServer) error {
	return nil
}
