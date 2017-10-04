package elastic

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/olivere/elastic.v5"
)

// TES provides the TES API endpoints, backed by elasticsearch.
type TES struct {
	*Elastic
	Backend compute.Backend
}

// NewTES creates a new TES API with the given config.
func NewTES(conf config.Elastic) (*TES, error) {
	es, err := NewElastic(conf)
	return &TES{Elastic: es}, err
}

// WithComputeBackend sets the compute backend.
func (et *TES) WithComputeBackend(b compute.Backend) {
	et.Backend = b
}

// CreateTask creates a new task.
func (et *TES) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {

	if err := tes.Validate(task); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	task.Id = util.GenTaskID()
	task.State = tes.State_QUEUED
	if err := et.Elastic.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	if et.Backend != nil {
		if err := et.Backend.Submit(task); err != nil {
			return nil, err
		}
	}
	return &tes.CreateTaskResponse{Id: task.Id}, nil
}

// GetTask gets a task by ID.
func (et *TES) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	resp, err := et.Elastic.GetTask(ctx, req.Id)
	if elastic.IsNotFound(err) {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: task ID: %s", err.Error(), req.Id))
	}
	return resp, err
}

func getPageSize(req *tes.ListTasksRequest) int {
	pageSize := 256

	if req.PageSize != 0 {
		pageSize = int(req.GetPageSize())
		if pageSize > 2048 {
			pageSize = 2048
		}
		if pageSize < 50 {
			pageSize = 50
		}
	}
	return pageSize
}

// ListTasks lists tasks.
func (et *TES) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	tasks, err := et.Elastic.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	return &tes.ListTasksResponse{Tasks: tasks}, nil
}

// CancelTask cancels a task by ID.
func (et *TES) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	return nil, nil
}

// GetServiceInfo returns service metadata.
func (et *TES) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{Name: "elastic"}, nil
}

// CreateEvent writes a task event to the database.
func (et *TES) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	return &events.CreateEventResponse{}, et.Elastic.Write(req)
}
