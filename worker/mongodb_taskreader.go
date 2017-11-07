package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/mongodb"
)

// MongoDBTaskReader provides read access to tasks from MongoDB
type MongoDBTaskReader struct {
	client *mongodb.MongoDB
	taskID string
}

// NewMongoDBTaskReader returns a new Mongo Task Reader.
func NewMongoDBTaskReader(conf config.MongoDB, taskID string) (*MongoDBTaskReader, error) {
	db, err := mongodb.NewMongoDB(conf)
	if err != nil {
		return nil, err
	}
	return &MongoDBTaskReader{db, taskID}, nil
}

// Task returns the task descriptor.
func (r *MongoDBTaskReader) Task() (*tes.Task, error) {
	return r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *MongoDBTaskReader) State() (tes.State, error) {
	t, err := r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}
