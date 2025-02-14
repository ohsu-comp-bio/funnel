package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	tes.UnimplementedTaskServiceServer
	Name    string
	Event   events.Writer
	Compute events.Computer
	Read    tes.ReadOnlyServer
	Log     *logger.Logger
	Config  config.Config
}

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (ts *TaskService) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {

	if err := tes.InitTask(task, true); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err.Error())
	}

	if err := ReplaceInputBearerToken(ctx, task); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := ts.Compute.CheckBackendParameterSupport(task)
	if err != nil {
		return nil, fmt.Errorf("error from backend: %s", err)
	}

	ctx = context.WithValue(ctx, "Config", ts.Config)
	if err := ts.Event.WriteEvent(ctx, events.NewTaskCreated(task)); err != nil {
		return nil, fmt.Errorf("error creating task: %s", err)
	}

	// dispatch to compute backend
	go func() {
		err := ts.Compute.WriteEvent(ctx, events.NewTaskCreated(task))
		if err != nil {
			ts.Log.Error("error submitting task to compute backend", "taskID", task.Id, "error", err)
		}
	}()

	return &tes.CreateTaskResponse{Id: task.Id}, nil
}

// GetTask calls GetTask on the underlying tes.ReadOnlyServer. If the underlying server
// returns tes.ErrNotFound, TaskService will handle returning the appropriate gRPC error.
func (ts *TaskService) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	task, err := ts.Read.GetTask(ctx, req)
	if err == tes.ErrNotFound {
		err = status.Errorf(codes.NotFound, "%v: taskID: %s", err.Error(), req.Id)
	}
	return task, err
}

// ListTasks calls ListTasks on the underlying tes.ReadOnlyServer.
func (ts *TaskService) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	return ts.Read.ListTasks(ctx, req)
}

// CancelTask cancels a task
func (ts *TaskService) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	result := &tes.CancelTaskResponse{}

	// updated database and other event streams (includes access-checking)
	err := ts.Event.WriteEvent(ctx, events.NewState(req.Id, tes.Canceled))
	if err == tes.ErrNotFound {
		return result, status.Errorf(codes.NotFound, "%v: taskID: %s", err.Error(), req.Id)
	} else if err == tes.ErrNotPermitted {
		return result, status.Errorf(codes.PermissionDenied, "%v: taskID: %s", err.Error(), req.Id)
	} else if err != nil {
		return result, err
	}

	// dispatch to compute backend
	err = ts.Compute.WriteEvent(ctx, events.NewState(req.Id, tes.Canceled))
	if err != nil {
		ts.Log.Error("compute backend failed to cancel task", "taskID", req.Id, "error", err)
	}

	return result, err
}

// GetServiceInfo returns service metadata.
func (ts *TaskService) GetServiceInfo(ctx context.Context, info *tes.GetServiceInfoRequest) (*tes.ServiceInfo, error) {
	resp := &tes.ServiceInfo{
		CreatedAt: "2016-03-21T16:27:49-07:00",
		// TODO: Change this to "mailto:ellrott@ohsu.edu" when support for "mailto:" URL's are
		// added to tes-compliance-suite
		ContactUrl:       "https://ohsu-comp-bio.github.io/funnel/",
		Description:      "Funnel is a toolkit for distributed task execution via a simple, standard API.",
		DocumentationUrl: "https://ohsu-comp-bio.github.io/funnel/",
		Environment:      "development",
		Id:               "org.ga4gh.funnel",
		Name:             ts.Name,
		Organization: map[string]string{
			"name": "OHSU Computational Biology",
			"url":  "https://github.com/ohsu-comp-bio",
		},
		Storage: []string{
			"file:///path/to/local/funnel-storage",
			"s3://ohsu-compbio-funnel/storage",
		},
		TesResourcesBackendParameters: []string{
			"",
		},
		Type: &tes.ServiceType{
			Artifact: "tes",
			Group:    "org.ga4gh",
			Version:  version.String(),
		},
		UpdatedAt: time.Now().Format(time.RFC3339),
		Version:   version.String(),
	}

	/*
		//Task metrics no longer in service info as of TES 1.1
		if c, ok := ts.Read.(metrics.TaskStateCounter); ok {
			resp.TaskStateCounts = make(map[string]int32)
			// Ensure that all states are present in the response, even if zero.
			for key := range tes.State_value {
				resp.TaskStateCounts[key] = 0
			}
			cs, err := c.TaskStateCounts(ctx)
			if err != nil {
				ts.Log.Error("counting task states", "error", err)
			}
			// Override the zero values in the response.
			for key, count := range cs {
				resp.TaskStateCounts[key] = count
			}
		}
	*/

	return resp, nil
}

func ReplaceInputBearerToken(ctx context.Context, task *tes.Task) error {
	userInfo, ok := ctx.Value(UserInfoKey).(*UserInfo)
	noToken := !ok || userInfo.Token == ""

	for _, input := range task.Inputs {
		if !strings.HasPrefix(input.Url, "sda://") &&
			!strings.HasPrefix(input.Url, "htsget://") ||
			strings.Contains(input.Url, "#") {
			continue
		}
		if noToken {
			scheme, _, _ := strings.Cut(input.Url, "://")
			if scheme == "htsget" {
				continue
			}
			return errors.New("Task input from SDA requires a Bearer token " +
				"to be used for fetching the data, however, current " +
				"authentication-context has no information about the token " +
				"to use. If necessary, please provide an explicit Bearer " +
				"token in the URL right after the hash-sign ('#'): " +
				"sda://dataset-id/file/path#bearer-token")
		}
		input.Url = input.Url + "#" + userInfo.Token
	}

	return nil
}
