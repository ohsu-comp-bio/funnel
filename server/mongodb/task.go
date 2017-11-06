package mongodb

import (
	// "fmt"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (db *MongoDB) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	verr := tes.Validate(task)
	if verr != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, verr.Error())
	}

	taskID := util.GenTaskID()
	task.Id = taskID
	task.State = tes.State_QUEUED

	// TODO write to database and submit

	return &tes.CreateTaskResponse{Id: taskID}, nil
}

// GetTask gets a task, which describes a running task
func (db *MongoDB) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task

	return task, nil
}

// ListTasks returns a list of taskIDs
func (db *MongoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	var tasks []*tes.Task
	var pageSize int64 = 256

	if req.PageSize != 0 {
		pageSize = int64(req.GetPageSize())
		if pageSize > 2048 {
			pageSize = 2048
		}
		if pageSize < 50 {
			pageSize = 50
		}
	}

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	return &out, nil
}

// CancelTask cancels a task
func (db *MongoDB) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	return &tes.CancelTaskResponse{}, nil
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
func (db *MongoDB) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{}, nil
}
