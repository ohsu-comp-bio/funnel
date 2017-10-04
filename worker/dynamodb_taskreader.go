package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/dynamodb"
)

// DynamoDBTaskReader provides read access to tasks from DynamoDB
type DynamoDBTaskReader struct {
	client *dynamodb.DynamoDB
	taskID string
}

// NewDynamoDBTaskReader returns a new reader.
func NewDynamoDBTaskReader(conf config.DynamoDB, taskID string) (*DynamoDBTaskReader, error) {
	db, err := dynamodb.NewDynamoDB(conf)
	if err != nil {
		return nil, err
	}
	return &DynamoDBTaskReader{db, taskID}, nil
}

// Task returns the task descriptor.
func (r *DynamoDBTaskReader) Task() (*tes.Task, error) {
	return r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *DynamoDBTaskReader) State() (tes.State, error) {
	t, err := r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}
