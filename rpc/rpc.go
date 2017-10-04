package rpc

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
	"time"
)

// TESClient provides read access to tasks from the funnel server over gRPC.
type TESClient struct {
	tes.TaskServiceClient
}

// NewTESClient returns a new TES RPC client with the given configuration,
// including transport security, basic password auth, and dial timeout.
func NewTESClient(conf config.RPC) (*TESClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		conf.ServerAddress,
		grpc.WithInsecure(),
		util.PerRPCPassword(conf.ServerPassword),
	)
	if err != nil {
		return nil, err
	}
	cli := tes.NewTaskServiceClient(conn)

	return &TESClient{cli}, nil
}

// FullTask returns the task descriptor.
func (r *TESClient) FullTask(id string) (*tes.Task, error) {
	return r.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *TESClient) State(id string) (tes.State, error) {
	t, err := r.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}
