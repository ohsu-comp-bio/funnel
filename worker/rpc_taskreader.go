package worker

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/tes"
	util "github.com/ohsu-comp-bio/funnel/util/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// RPCTaskReader provides read access to tasks from the funnel server over gRPC.
type RPCTaskReader struct {
	client tes.TaskServiceClient
	conn   *grpc.ClientConn
}

// NewRPCTaskReader returns a new RPC-based task reader.
func NewRPCTaskReader(ctx context.Context, conf config.Server) (*RPCTaskReader, error) {
	conn, err := util.Dial(ctx, conf)
	if err != nil {
		return nil, err
	}
	cli := tes.NewTaskServiceClient(conn)
	return &RPCTaskReader{cli, conn}, nil
}

// Task returns the task descriptor.
func (r *RPCTaskReader) Task(ctx context.Context, taskID string) (*tes.Task, error) {
	return r.client.GetTask(ctx, &tes.GetTaskRequest{
		Id:   taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *RPCTaskReader) State(ctx context.Context, taskID string) (tes.State, error) {
	t, err := r.client.GetTask(ctx, &tes.GetTaskRequest{
		Id:   taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}

// Close closes the connection.
func (r *RPCTaskReader) Close() {
	r.conn.Close()
}
