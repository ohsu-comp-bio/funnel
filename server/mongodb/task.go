package mongodb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var basicView = bson.M{"logs.logs.stdout": 0, "logs.logs.stderr": 0, "inputs.contents": 0}
var minimalView = bson.M{"id": 1, "state": 1}

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

	task.Logs = []*tes.TaskLog{
		{
			Logs: []*tes.ExecutorLog{},
		},
	}

	err := db.tasks.Insert(task)
	if err != nil {
		return nil, fmt.Errorf("failed to write task to db: %v", err)
	}

	err = db.backend.Submit(task)
	if err != nil {
		return nil, fmt.Errorf("couldn't submit to compute backend: %v", err)
	}

	return &tes.CreateTaskResponse{Id: taskID}, nil
}

// GetTask gets a task, which describes a running task
func (db *MongoDB) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task tes.Task
	var q *mgo.Query

	q = db.tasks.Find(bson.M{"id": req.Id})
	switch req.View {
	case tes.TaskView_BASIC:
		q = q.Select(basicView)
	case tes.TaskView_MINIMAL:
		q = q.Select(minimalView)
	}

	err := q.One(&task)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", mgo.ErrNotFound.Error(), req.Id))
		}
		return nil, err
	}

	return &task, nil
}

// ListTasks returns a list of taskIDs
func (db *MongoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	pageSize := tes.GetPageSize(req.GetPageSize())

	var q *mgo.Query
	var err error
	if req.PageToken != "" {
		q = db.tasks.Find(bson.M{"id": bson.M{"$gt": req.PageToken}})
	} else {
		q = db.tasks.Find(nil)
	}
	q = q.Limit(pageSize)

	switch req.View {
	case tes.TaskView_BASIC:
		q = q.Select(basicView)
	case tes.TaskView_MINIMAL:
		q = q.Select(minimalView)
	}

	var tasks []*tes.Task
	err = q.All(&tasks)
	if err != nil {
		return nil, err
	}

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}
	// TODO figure out when not to return a next page token
	if len(tasks) > 0 {
		out.NextPageToken = tasks[len(tasks)-1].Id
	}

	return &out, nil
}

// CancelTask cancels a task
func (db *MongoDB) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	task, err := db.GetTask(ctx, &tes.GetTaskRequest{
		Id:   req.Id,
		View: tes.TaskView_MINIMAL,
	})
	if err != nil {
		return nil, err
	}

	from := task.State
	to := tes.State_CANCELED
	if err := tes.ValidateTransition(from, to); err != nil {
		return nil, err
	}

	err = db.tasks.Update(bson.M{"id": req.Id}, bson.M{"$set": bson.M{"state": to}})

	return &tes.CancelTaskResponse{}, err
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
func (db *MongoDB) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{}, nil
}
