package mongodb

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/tes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
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
	var opts = options.FindOne()

	switch req.View {
	case tes.View_BASIC.String():
		opts = opts.SetProjection(basicView)
	case tes.View_MINIMAL.String():
		opts = opts.SetProjection(minimalView)
	}

	err := db.tasks(db.client).FindOne(context.TODO(), bson.M{"id": req.Id}, opts).Decode(&task)
	if err != nil {
		return nil, tes.ErrNotFound
	}

	return &task, nil
}

// ListTasks returns a list of taskIDs
func (db *MongoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	pageSize := tes.GetPageSize(req.GetPageSize())

	var query = bson.M{}
	var err error
	if req.PageToken != "" {
		query["id"] = bson.M{"$lt": req.PageToken}
	}

	if req.State != tes.Unknown {
		query["state"] = bson.M{"$eq": req.State}
	}

	if req.NamePrefix != "" {
		query["name"] = bson.M{"$regex": fmt.Sprintf("^%s", req.NamePrefix)}
	}

	for k, v := range req.GetTags() {
		if v == "" {
			query[fmt.Sprintf("tags.%s", k)] = bson.M{"$exists": true}
		} else {
			query[fmt.Sprintf("tags.%s", k)] = bson.M{"$eq": v}
		}
	}

	var opts = options.Find().SetSort(bson.M{"creationtime": -1}).SetLimit(int64(pageSize))

	switch req.View {
	case tes.View_BASIC.String():
		opts = opts.SetProjection(basicView)
	case tes.View_MINIMAL.String():
		opts = opts.SetProjection(minimalView)
	}

	cursor, err := db.tasks(db.client).Find(context.TODO(), query, opts)
	if err != nil {
		return nil, err
	}

	var tasks []*tes.Task
	err = cursor.All(context.TODO(), &tasks)
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
