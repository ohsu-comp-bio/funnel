package worker

import (
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// GenericTaskReader provides read access to tasks.
type GenericTaskReader struct {
	get    func(ctx context.Context, in *tes.GetTaskRequest) (*tes.Task, error)
	close  func()
	taskID string
}

// NewGenericTaskReader returns a new generic task reader.
func NewGenericTaskReader(get func(ctx context.Context, in *tes.GetTaskRequest) (*tes.Task, error), taskID string, close func()) *GenericTaskReader {
	return &GenericTaskReader{get, close, taskID}
}

// Task returns the task descriptor.
func (r *GenericTaskReader) Task(ctx context.Context) (*tes.Task, error) {
	return r.get(ctx, &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *GenericTaskReader) State(ctx context.Context) (tes.State, error) {
	t, err := r.get(ctx, &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_MINIMAL,
	})
	return t.GetState(), err
}

func (r *GenericTaskReader) Close() {
	r.close()
}