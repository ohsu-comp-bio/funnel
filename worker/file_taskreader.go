package worker

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// FileTaskReader provides a TaskReader implementation from a task file.
type FileTaskReader struct {
	Path string
	task *tes.Task
}

// Task returns the task. A random ID will be generated.
func (f *FileTaskReader) Task(ctx context.Context) (*tes.Task, error) {
	if f.task != nil {
		return f.task, nil
	}

	err := f.load()
	if err != nil {
		return nil, err
	}

	return f.task, nil
}

// State returns the task state. Due to some quirks in the implementation
// of this reader, this will always return QUEUED.
func (f *FileTaskReader) State(ctx context.Context) (tes.State, error) {
	// TODO tes.InitTask is being used internally, which initializes the
	//      task to QUEUED. So this will always return queued.

	if f.task != nil {
		return f.task.State, nil
	}

	err := f.load()
	if err != nil {
		return tes.Unknown, err
	}

	return f.task.State, nil
}

func (f *FileTaskReader) load() error {
	fh, err := os.Open(f.Path)
	if err != nil {
		return fmt.Errorf("opening task file: %v", err)
	}
	defer fh.Close()

	task := &tes.Task{}
	err = jsonpb.Unmarshal(fh, task)
	if err != nil {
		return fmt.Errorf("unmarshaling task file: %v", err)
	}
	err = tes.InitTask(task)
	if err != nil {
		return fmt.Errorf("validating task: %v", err)
	}

	f.task = task
	return nil
}
