package worker

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

// GenericTaskReader provides read access to tasks.
type GenericTaskReader struct {
	get func(ctx context.Context, in *tes.GetTaskRequest) (*tes.Task, error)
}

// NewGenericTaskReader returns a new generic task reader.
func NewGenericTaskReader(get func(ctx context.Context, in *tes.GetTaskRequest) (*tes.Task, error)) *GenericTaskReader {
	return &GenericTaskReader{get}
}

// Task returns the task descriptor.
func (r *GenericTaskReader) Task(ctx context.Context, taskID string) (*tes.Task, error) {
	return r.get(ctx, &tes.GetTaskRequest{
		Id:   taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *GenericTaskReader) State(ctx context.Context, taskID string) (tes.State, error) {
	t, err := r.get(ctx, &tes.GetTaskRequest{
		Id:   taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}
