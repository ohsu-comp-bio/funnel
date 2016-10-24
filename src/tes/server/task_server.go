package tes_server

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"tes/ga4gh"
	"tes/server/proto"
)

// GA4GHServer that is common
// Multiple services could be placed into the same server
// For the moment there is just the task server
// TODO: documentation
type GA4GHServer struct {
	task  ga4gh_task_exec.TaskServiceServer
	sched ga4gh_task_ref.SchedulerServer
}

// NewGA4GHServer documentation
// TODO: documentation
func NewGA4GHServer() *GA4GHServer {
	return &GA4GHServer{}
}

// RegisterTaskServer documentation
// TODO: documentation
func (ga4ghServer *GA4GHServer) RegisterTaskServer(task ga4gh_task_exec.TaskServiceServer) {
	ga4ghServer.task = task
}

// RegisterScheduleServer documentation
// TODO: documentation
func (ga4ghServer *GA4GHServer) RegisterScheduleServer(sched ga4gh_task_ref.SchedulerServer) {
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
		ga4gh_task_exec.RegisterTaskServiceServer(grpcServer, ga4ghServer.task)
	}
	if ga4ghServer.sched != nil {
		ga4gh_task_ref.RegisterSchedulerServer(grpcServer, ga4ghServer.sched)
	}

	log.Println("Starting RPC Server on " + hostPort)
	go grpcServer.Serve(lis)
}
