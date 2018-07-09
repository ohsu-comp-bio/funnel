package worker

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type Base64TaskReader struct {
	task *tes.Task
}

func NewBase64TaskReader(raw string) (*Base64TaskReader, error) {
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decoding task: %v", err)
	}

	task := &tes.Task{}
	buf := bytes.NewBuffer(data)
	err = jsonpb.Unmarshal(buf, task)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling task: %v", err)
	}

	err = tes.InitTask(task, false)
	if err != nil {
		return nil, fmt.Errorf("initializing task: %v", err)
	}

	return &Base64TaskReader{task}, nil
}

// Task returns the task. A random ID will be generated.
func (f *Base64TaskReader) Task(ctx context.Context) (*tes.Task, error) {
	return f.task, nil
}

// State returns the task state. Due to some quirks in the implementation
// of this reader, and since there is no online database to connect to,
// this will always return QUEUED.
func (f *Base64TaskReader) State(ctx context.Context) (tes.State, error) {
	return f.task.GetState(), nil
}
