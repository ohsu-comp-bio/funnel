package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
	"time"
)

// RPCTaskReader provides read access to tasks from the funnel server over gRPC.
type RPCTaskReader struct {
	client tes.TaskServiceClient
	taskID string
}

// NewRPCTaskReader returns a new RPC-based task reader.
func NewRPCTaskReader(conf config.Worker, taskID string) (*RPCTaskReader, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		conf.EventWriters.RPC.ServerAddress,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		util.PerRPCPassword(conf.EventWriters.RPC.ServerPassword),
	)
	if err != nil {
		return nil, err
	}
	cli := tes.NewTaskServiceClient(conn)

	return &RPCTaskReader{cli, taskID}, nil
}

// Task returns the task descriptor.
func (r *RPCTaskReader) Task() (*tes.Task, error) {
	return r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *RPCTaskReader) State() (tes.State, error) {
	t, err := r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}
