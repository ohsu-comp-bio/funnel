package mongodb

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var basicView = bson.M{
	"logs.systemlogs":  0,
	"logs.logs.stdout": 0,
	"logs.logs.stderr": 0,
	"inputs.content":   0,
}
var minimalView = bson.M{"id": 1, "state": 1}

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
	if err == mgo.ErrNotFound {
		return nil, tes.ErrNotFound
	}
	if err != nil {
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
