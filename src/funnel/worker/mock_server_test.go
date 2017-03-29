package worker

import (
	"funnel/config"
	tes "funnel/proto/tes"
	"funnel/scheduler"
	"funnel/server"
	server_mocks "funnel/server/mocks"
	pbf "funnel/proto/funnel"
)

// mockScheduler is a mock scheduler that assigns every job to a single worker.
type mockScheduler struct {
	worker *pbf.Worker
}

func (m *mockScheduler) Schedule(j *tes.Job) *scheduler.Offer {
	return scheduler.NewOffer(m.worker, j, scheduler.Scores{})
}

func newMockSchedulerServer() *MockSchedulerServer {
	conf := config.DefaultConfig()
	return MockSchedulerServerFromConfig(conf)
}

// MockSchedulerServerFromConfig returns a mockSchdulerServer with the given config.
func MockSchedulerServerFromConfig(conf config.Config) *MockSchedulerServer {
	srv := server_mocks.MockServerFromConfig(conf)

	conf.Worker.ServerAddress = srv.Conf.HostName + ":" + srv.Conf.RPCPort
	conf.Worker.ID = "test-worker"

	// Create a worker
	w, werr := NewWorker(conf.Worker)
	if werr != nil {
		panic(werr)
	}
	// Stub the job runner so it's a no-op runner
	// i.e. ensure docker run, file copying, etc. doesn't actually happen
	w.JobRunner = NoopJobRunner

	// Create a mock scheduler with a single worker
	sched := &mockScheduler{&pbf.Worker{
		Id:    "test-worker",
		State: pbf.WorkerState_Alive,
		Jobs:  map[string]*pbf.JobWrapper{},
	}}

	m := MockSchedulerServer{srv.DB, srv, sched, srv.Conf, w}
	return &m
}

type MockSchedulerServer struct {
	db     *server.TaskBolt
	Server *server_mocks.MockServer
	sched  *mockScheduler
	conf   config.Config
	worker *Worker
}

func (m *MockSchedulerServer) Flush() {
	scheduler.ScheduleChunk(m.db, m.sched, m.conf)
	m.worker.Sync()
}

func (m *MockSchedulerServer) Close() {
	m.Server.Close()
}
