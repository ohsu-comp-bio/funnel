package elastic

import (
  "golang.org/x/net/context"
  "github.com/ohsu-comp-bio/funnel/proto/tes"
  "github.com/ohsu-comp-bio/funnel/util"
  "github.com/ohsu-comp-bio/funnel/events"
  "github.com/ohsu-comp-bio/funnel/compute"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
)

type TES struct {
  *Elastic
  Backend compute.Backend
}

func NewTES(conf Config) (*TES, error) {
  es, err := NewElastic(conf)
  return &TES{Elastic: es}, err
}

func (et *TES) WithComputeBackend(b compute.Backend) {
  et.Backend = b
}

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
  return &tes.CreateTaskResponse{task.Id}, nil
}

func (et *TES) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
  return et.Elastic.GetTask(ctx, req.Id)
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

func (et *TES) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
  tasks, err := et.Elastic.ListTasks(ctx)
  if err != nil {
    return nil, err
  }
  return &tes.ListTasksResponse{Tasks: tasks}, nil
}

func (et *TES) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
  return nil, nil
}

func (et *TES) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
  return &tes.ServiceInfo{Name: "elastic"}, nil
}

func (et *TES) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
  return &events.CreateEventResponse{}, et.Elastic.Write(req)
}
