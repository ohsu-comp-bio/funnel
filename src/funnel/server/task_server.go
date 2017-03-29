package server

import (
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"google.golang.org/grpc"
	"net"
)

// GA4GHServer that is common. While multiple services could be
// placed into the same server, for the moment there is just the task
// server.
type GA4GHServer struct {
	task  tes.TaskServiceServer
	sched pbf.SchedulerServer
}

// NewGA4GHServer documentation
// TODO: documentation
func NewGA4GHServer() *GA4GHServer {
	return &GA4GHServer{}
}

// RegisterTaskServer documentation
// TODO: documentation
func (ga4ghServer *GA4GHServer) RegisterTaskServer(task tes.TaskServiceServer) {
	ga4ghServer.task = task
}

// RegisterScheduleServer documentation
// TODO: documentation
func (ga4ghServer *GA4GHServer) RegisterScheduleServer(sched pbf.SchedulerServer) {
	ga4ghServer.sched = sched
}

// Start documentation
// TODO: documentation
func (ga4ghServer *GA4GHServer) Start(hostPort string) {
	lis, err := net.Listen("tcp", ":"+hostPort)
	if err != nil {
		panic("Cannot open port")
	}
	grpcServer := grpc.NewServer()

	if ga4ghServer.task != nil {
		tes.RegisterTaskServiceServer(grpcServer, ga4ghServer.task)
	}
	if ga4ghServer.sched != nil {
		pbf.RegisterSchedulerServer(grpcServer, ga4ghServer.sched)
	}

	log.Info("RPC server listening", "port", hostPort)
	grpcServer.Serve(lis)
}
