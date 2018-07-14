package papi

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/api/genomics/v2alpha1"
)

// Submit submits an operation to the Pipelines service.
func (b *Backend) Submit(task *tes.Task) error {
	ctx := context.Background()
	papiID, err := b.submit(task)
	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to Google Pipelines",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

	b.event.WriteEvent(
		ctx,
		events.NewSystemLog(
			task.Id, 0, 0, "info",
			"submitted task to Google Pipelines", nil,
		),
	)

	return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{
			"pipeline_operation_id": papiID,
		}),
	)
}

func (b *Backend) submit(task *tes.Task) (string, error) {

	res, err := getResources(b.conf, task)
	if err != nil {
		return "", fmt.Errorf("determining resources: %v", err)
	}

	// encode the task as a string so that it can be sent
	// as part of the pipelines API call.
	data := &bytes.Buffer{}
	mar := jsonpb.Marshaler{}
	_ = mar.Marshal(data, task)
	b64 := base64.StdEncoding.EncodeToString(data.Bytes())

	args := []string{"worker", "run", "--taskBase64", b64}
	args = append(args, b.conf.ExtraArgs...)

	pl := &genomics.Pipeline{
		Environment: map[string]string{
			"FUNNEL_TASK_ID": task.Id,
			"FUNNEL_TASK":    b64,
		},
		Resources: res,
		Actions: []*genomics.Action{
			{
				Name:     fmt.Sprintf("funnel-task-%s", task.Id),
				Commands: args,
				ImageUri: b.conf.WorkerImage,
			},
		},
	}

	call := b.client.Pipelines.Run(&genomics.RunPipelineRequest{
		Pipeline: pl,
	})

	resp, err := call.Do()
	return resp.Name, err
}

// Cancel cancels a running Pipelines Operation.
func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	task, err := b.database.GetTask(
		ctx, &tes.GetTaskRequest{Id: taskID, View: tes.TaskView_BASIC},
	)
	if err != nil {
		return fmt.Errorf("getting task: %v", err)
	}

	if task.State == tes.Canceled {
		return nil
	}
	if tes.TerminalState(task.State) {
		return fmt.Errorf("task is already in a terminal state %q", task.State)
	}

	jobid, ok := task.GetTaskLog(0).GetMetadata()["pipeline_operation_id"]
	if !ok {
		return fmt.Errorf("no Google Pipelines operation ID found in metadata for task %s", taskID)
	}

	// TODO technically, Google Pipelines might fail to cancel the task.
	//      This might be better suited for a reconciler check.
	req := &genomics.CancelOperationRequest{}
	_, err = b.client.Projects.Operations.Cancel(jobid, req).Context(ctx).Do()
	return err
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ctx, ev.Id)
		}
	}
	return nil
}
