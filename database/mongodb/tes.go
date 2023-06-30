package mongodb

import (
	"encoding/json"
	"fmt"

	"github.com/globalsign/mgo/bson"
	"github.com/ohsu-comp-bio/funnel/tes"
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
	// return nil, errors.New("not implemented")
	// sess := db.sess.Copy()
	// defer sess.Close()

	sess := db.client
	defer sess.Disconnect(context.TODO())

	var task tes.Task
	// var q *mgo.Query

	cursor, err := db.tasks(sess).Find(context.TODO(), bson.M{"id": req.Id})
	if err = cursor.All(context.TODO(), &task); err != nil {
		return nil, err
	}
	res, _ := json.Marshal(task)
	fmt.Println(string(res))
	return &task, nil

	// switch req.View {
	// case tes.View_BASIC.String():
	// 	cursor = cursor.Select(basicView)
	// case tes.View_MINIMAL.String():
	// 	cursor = cursor.Select(minimalView)
	// }

	// err = cursor.One(&task)
	// if err == mgo.ErrNotFound {
	// 	return nil, tes.ErrNotFound
	// }
	// if err != nil {
	// 	return nil, err
	// }

	// return &task, nil
}

// ListTasks returns a list of taskIDs
func (db *MongoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	// return nil, errors.New("not implemented")
	// sess := db.sess.Copy()
	// defer sess.Close()

	sess := db.client
	defer sess.Disconnect(context.TODO())

	// pageSize := tes.GetPageSize(req.GetPageSize())

	var query = bson.M{}
	// var q *mgo.Query
	var err error
	if req.PageToken != "" {
		query["id"] = bson.M{"$lt": req.PageToken}
	}

	if req.State != tes.Unknown {
		query["state"] = bson.M{"$eq": req.State}
	}

	for k, v := range req.GetTags() {
		query[fmt.Sprintf("tags.%s", k)] = bson.M{"$eq": v}
	}

	cursor, err := db.tasks(sess).Find(context.TODO(), bson.M{})//.Sort("-creationtime").Limit(pageSize)
	if err != nil {
		return nil, err
	}

	// switch req.View {
	// case tes.View_BASIC.String():
	// 	q = q.Select(basicView)
	// case tes.View_MINIMAL.String():
	// 	q = q.Select(minimalView)
	// }

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
