package server

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// TaskService is a wrapper which handles common TES Task Service operations,
// such as initializing a task when CreateTask is called. The TaskService is backed by
// two parts: a read API which provides the GetTask and ListTasks endpoints, and a write
// API which implements the events.Writer interface. Task creation and cancelation is
// managed by writing events to underlying event writer.
//
// This makes it easier to define task service backends for new databases, and ensures
// that common operations are handled consistently, such as setting IDs, handling 404s,
// GetServiceInfo, etc.
type TaskService struct {
	Name  string
	Event events.Writer
	Read  tes.ReadOnlyServer
}

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (ts *TaskService) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {

	if err := tes.InitTask(task); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}
	err := ts.Event.WriteEvent(ctx, events.NewTaskCreated(task))

	if err != nil {
		ts.Event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		return nil, fmt.Errorf("can't create task: %s", err)
	}

	return &tes.CreateTaskResponse{Id: task.Id}, nil
}

// GetTask calls GetTask on the underlying tes.ReadOnlyServer. If the underlying server
// returns tes.ErrNotFound, TaskService will handle returning the appropriate gRPC error.
func (ts *TaskService) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	task, err := ts.Read.GetTask(ctx, req)
	if err == tes.ErrNotFound {
		err = grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", err.Error(), req.Id))
	}
	return task, err
}

// ListTasks calls ListTasks on the underlying tes.ReadOnlyServer.
func (ts *TaskService) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	return ts.Read.ListTasks(ctx, req)
}

// CancelTask cancels a task
func (ts *TaskService) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	err := ts.Event.WriteEvent(ctx, events.NewState(req.Id, tes.Canceled))
	if err == tes.ErrNotFound {
		err = grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", err.Error(), req.Id))
	}
	return &tes.CancelTaskResponse{}, err
}

// GetServiceInfo returns service metadata.
func (ts *TaskService) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{Name: ts.Name, Doc: version.String()}, nil
}
