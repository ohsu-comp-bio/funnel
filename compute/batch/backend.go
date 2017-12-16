package batch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
	"regexp"
	"time"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, batchConf config.AWSBatch, reader tes.ReadOnlyServer, writer events.Writer) (*Backend, error) {

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

	go b.reconcile(ctx)

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
		return b.Submit(ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ev.Id)
		}
	}
	return nil
}

// Submit submits a task to the AWS batch service.
func (b *Backend) Submit(task *tes.Task) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	resp, err := b.client.SubmitJobWithContext(ctx, req)
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
func (b *Backend) Cancel(taskID string) error {
	var task *tes.Task
	var err error

	task, err = b.database.GetTask(
		context.Background(), &tes.GetTaskRequest{Id: taskID, View: tes.TaskView_FULL},
	)
	if err != nil {
		return err
	}

	// only cancel tasks in a QUEUED state
	state := task.GetState()
	if state != tes.State_QUEUED {
		return nil
	}

	backendID := getAWSTaskID(task)
	_, err = b.client.CancelJob(&batch.CancelJobInput{
		JobId:  aws.String(backendID),
		Reason: aws.String("User requested cancel"),
	})

	return err
}

// Reconcile loops through tasks and checks the status from Funnel's database
// against the status reported by AWS Batch. This allows the backend to report
// system error's that prevented the worker process from running.
//
// Currently this handles a narrow set of cases:
//
// |---------------------|-----------------|--------------------|
// |    Funnel State     |  Backend State  |  Reconciled State  |
// |---------------------|-----------------|--------------------|
// |        QUEUED       |     FAILED      |    SYSTEM_ERROR    |
// |  INITIALIZING       |     FAILED      |    SYSTEM_ERROR    |
// |       RUNNING       |     FAILED      |    SYSTEM_ERROR    |
//
// In this context a "FAILED" state is being used as a generic term that captures
// one or more terminal states for the backend.
func (b *Backend) reconcile(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pageToken := ""
			for {
				lresp, _ := b.database.ListTasks(ctx, &tes.ListTasksRequest{
					View:      tes.TaskView_BASIC,
					PageSize:  100,
					PageToken: pageToken,
				})
				pageToken = lresp.NextPageToken

				tmap := make(map[string]*tes.Task)
				var jobs []*string
				for _, t := range lresp.Tasks {
					switch t.State {
					case tes.Queued, tes.Initializing, tes.Running:
						jobid := getAWSTaskID(t)
						tmap[jobid] = t
						jobs = append(jobs, aws.String(jobid))
					}
				}

				resp, _ := b.client.DescribeJobs(&batch.DescribeJobsInput{
					Jobs: jobs,
				})

				for _, j := range resp.Jobs {
					task := tmap[*j.JobId]
					jstate := *j.Status

					if jstate == "FAILED" {
						b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
						b.event.WriteEvent(
							ctx,
							events.NewSystemLog(
								task.Id, 0, 0, "error",
								"AWSBatch job in FAILED state",
								map[string]string{"error": *j.StatusReason, "awsbatch_id": *j.JobId},
							),
						)
					}
				}

				// continue to next page from ListTasks or break
				if pageToken == "" {
					break
				}
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
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
		return metadata["awsbatch_id"]
	}
	return ""
}
