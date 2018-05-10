// Package batch contains code for accessing compute resources via AWS Batch.
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
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, conf config.AWSBatch, reader tes.ReadOnlyServer, writer events.Writer) (*Backend, error) {
	sess, err := util.NewAWSSession(conf.AWSConfig)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating batch client: %v", err)
	}

	b := &Backend{
		client:   batch.New(sess),
		conf:     conf,
		event:    writer,
		database: reader,
	}

	if !conf.DisableReconciler {
		go b.reconcile(ctx)
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
		ContainerOverrides: &batch.ContainerOverrides{},
	}

	// convert ram from GB to MiB
	if task.Resources != nil {
		ram := int64(task.Resources.RamGb * 953.674)
		if ram > 0 {
			req.ContainerOverrides.Memory = aws.Int64(ram)
		}

		vcpus := int64(task.Resources.CpuCores)
		if vcpus > 0 {
			req.ContainerOverrides.Vcpus = aws.Int64(vcpus)
		}
	}

	resp, err := b.client.SubmitJob(req)
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
	ticker := time.NewTicker(time.Duration(b.conf.ReconcileRate))

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pageToken := ""
			states := []tes.State{tes.Queued, tes.Initializing, tes.Running}
			for _, s := range states {
				for {
					lresp, _ := b.database.ListTasks(ctx, &tes.ListTasksRequest{
						View:      tes.TaskView_BASIC,
						State:     s,
						PageSize:  100,
						PageToken: pageToken,
					})
					pageToken = lresp.NextPageToken

					tmap := make(map[string]*tes.Task)
					var jobs []*string
					for _, t := range lresp.Tasks {
						jobid := getAWSTaskID(t)
						if jobid != "" {
							tmap[jobid] = t
							jobs = append(jobs, aws.String(jobid))
						}
					}

					if len(jobs) == 0 {
						if pageToken == "" {
							break
						}
						continue
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
						logger.Debug("End reconcile")
						break
					}
					time.Sleep(time.Millisecond * 100)
				}
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
		if metadata != nil {
			return metadata["awsbatch_id"]
		}
	}
	return ""
}
