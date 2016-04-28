
package ga4gh_task

import (
	"log"
	"net"
	"google.golang.org/grpc"
	"ga4gh-tasks"
)

/// Common GA4GH server, multiple services could be placed into the same server
/// For the moment there is just the task server
type GA4GHServer struct {
	task ga4gh_task_exec.TaskServiceServer
}

func NewGA4GHServer() *GA4GHServer {
	return &GA4GHServer {}
}

func (self *GA4GHServer) RegisterTaskServer(task ga4gh_task_exec.TaskServiceServer) {
	self.task = task
}


func (self *GA4GHServer) Run(host_port string) {
	lis, err := net.Listen("tcp", host_port)
	if err != nil {
		panic("Cannot open port")
	}
	grpcServer := grpc.NewServer()

	if self.task != nil {
		ga4gh_task_exec.RegisterTaskServiceServer(grpcServer, self.task)
	}
	log.Println("Starting Server")
	grpcServer.Serve(lis)
}
