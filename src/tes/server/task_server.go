package ga4gh_task

import (
	"tes/server/proto"
	"tes/ga4gh"
	"google.golang.org/grpc"
	"log"
	"net"
)

/// Common GA4GH server, multiple services could be placed into the same server
/// For the moment there is just the task server
type GA4GHServer struct {
	task  ga4gh_task_exec.TaskServiceServer
	sched ga4gh_task_ref.SchedulerServer
}

func NewGA4GHServer() *GA4GHServer {
	return &GA4GHServer{}
}

func (self *GA4GHServer) RegisterTaskServer(task ga4gh_task_exec.TaskServiceServer) {
	self.task = task
}

func (self *GA4GHServer) RegisterScheduleServer(sched ga4gh_task_ref.SchedulerServer) {
	self.sched = sched
}

func (self *GA4GHServer) Start(host_port string) {
	lis, err := net.Listen("tcp", ":"+host_port)
	if err != nil {
		panic("Cannot open port")
	}
	grpcServer := grpc.NewServer()

	if self.task != nil {
		ga4gh_task_exec.RegisterTaskServiceServer(grpcServer, self.task)
	}
	if self.sched != nil {
		ga4gh_task_ref.RegisterSchedulerServer(grpcServer, self.sched)
	}

	log.Println("Starting RPC Server on " + host_port)
	go grpcServer.Serve(lis)
}
