package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	util "github.com/ohsu-comp-bio/funnel/util/rpc"
	"google.golang.org/grpc"
)

// RPCTaskReader provides read access to tasks from the funnel server over gRPC.
type RPCTaskReader struct {
	client tes.TaskServiceClient
	conn   *grpc.ClientConn
	taskID string
}

// NewRPCTaskReader returns a new RPC-based task reader.
func NewRPCTaskReader(ctx context.Context, conf config.Server, taskID string) (*RPCTaskReader, error) {
	conn, err := util.Dial(ctx, conf, grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	cli := tes.NewTaskServiceClient(conn)
	return &RPCTaskReader{cli, conn, taskID}, nil
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

// Close closes the connection.
func (r *RPCTaskReader) Close() {
	r.conn.Close()
}
