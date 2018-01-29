package batch

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
)

// NewBackend returns a new local Backend instance.
func NewBackend(batchConf config.AWSBatch, reader tes.ReadOnlyServer, writer events.Writer) (*Backend, error) {

	sess, err := util.NewAWSSession(batchConf.AWSConfig)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating batch client: %v", err)
	}

	b := &Backend{
		client:   batch.New(sess),
		conf:     batchConf,
		event:    writer,
		database: reader,
	}

	return b, nil
}

// Backend represents the local backend.
type Backend struct {
	client   *batch.Batch
	conf     config.AWSBatch
	event    events.Writer
	database tes.ReadOnlyServer
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ctx, ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ctx, ev.Id)
		}
	}
	return nil
}

// Submit submits a task to the AWS batch service.
func (b *Backend) Submit(ctx context.Context, task *tes.Task) error {
	req := &batch.SubmitJobInput{
		// JobDefinition: aws.String(b.jobDef),
		JobDefinition: aws.String(b.conf.JobDefinition),
		JobName:       aws.String(safeJobName(task.Name)),
		JobQueue:      aws.String(b.conf.JobQueue),
		Parameters: map[string]*string{
			// Include the taskID in the job parameters. This gets used by
			// the funnel 'worker run' cmd.
			"taskID": aws.String(task.Id),
		},
	}

	// convert ram from GB to MiB
	ram := int64(task.Resources.RamGb * 953.674)
	vcpus := int64(task.Resources.CpuCores)
	if ram > 0 {
		req.ContainerOverrides.Memory = aws.Int64(ram)
	}

	if vcpus > 0 {
		req.ContainerOverrides.Vcpus = aws.Int64(vcpus)
	}

	reqctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	resp, err := b.client.SubmitJobWithContext(reqctx, req)
	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to AWSBatch",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

	return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{"awsbatch_id": *resp.JobId}),
	)
}

// Cancel removes tasks from the AWS batch job queue.
func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	task, err := b.database.GetTask(
		ctx, &tes.GetTaskRequest{Id: taskID, View: tes.TaskView_BASIC},
	)
	if err != nil {
		return err
	}

	// only cancel tasks in a QUEUED state
	if task.State != tes.State_QUEUED {
		return nil
	}

	backendID := getAWSTaskID(task)
	if backendID == "" {
		return fmt.Errorf("no AWS Batch ID found in metadata for task %s", taskID)
	}

	_, err = b.client.CancelJob(&batch.CancelJobInput{
		JobId:  aws.String(backendID),
		Reason: aws.String("User requested cancel"),
	})
	return err
}

// AWS limits the characters allowed in job names,
// so replace invalid characters with underscores.
func safeJobName(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	return re.ReplaceAllString(s, "_")
}

func getAWSTaskID(task *tes.Task) string {
	logs := task.GetLogs()
	if len(logs) > 0 {
		metadata := logs[0].GetMetadata()
		if metadata != nil {
			return metadata["awsbatch_id"]
		}
	}
	return ""
}
