package postgres

import (
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// GetTask gets a task, which describes a running task
func (db *Postgres) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	// TODO: Implement

	return nil, nil
}

// ListTasks returns a list of taskIDs
func (db *Postgres) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	// TODO: Implement

	return nil, nil
}
