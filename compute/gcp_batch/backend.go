// Package gcp_batch contains code for accessing compute resources via Google Batch.
// ref: https://cloud.google.com/batch/docs
// ref: https://cloud.google.com/batch/docs/reference/rest
package gcp_batch

import (
	"context"
	"fmt"
	"strings"
	"time"

	batch "cloud.google.com/go/batch/apiv1"
	"cloud.google.com/go/batch/apiv1/batchpb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type Backend struct {
	client   client
	conf     *config.GCPBatch
	event    events.Writer
	database tes.ReadOnlyServer
	log      *logger.Logger
	events.Backend
}

func NewBackend(ctx context.Context, conf *config.GCPBatch, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	client, err := batch.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		client:   client,
		conf:     conf,
		event:    writer,
		database: reader,
		log:      log,
	}

	if !conf.DisableReconciler {
		go b.reconcile(ctx)
	}

	return b, nil
}

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

func (b *Backend) Close() {
	b.database.Close()
	b.event.Close()
}

// Example storage interface
func (b *Backend) NewStorage(conf config.Config) (*storage.GoogleCloud, error) {
	gs, nerr := storage.NewGoogleCloud(conf.GoogleStorage)
	if nerr != nil {
		return nil, fmt.Errorf("failed to configure Google Storage backend: %s", nerr)
	}

	return gs, nil
}

func (b *Backend) Submit(task *tes.Task) error {
	ctx := context.Background()

	// 1. Identify all unique buckets used by the task
	buckets := make(map[string]bool)
	extractBucketName := func(url string) string {
		if strings.HasPrefix(url, "gs://") {
			parts := strings.SplitN(strings.TrimPrefix(url, "gs://"), "/", 2)
			return parts[0]
		}
		return ""
	}

	for _, input := range task.Inputs {
		if bucket := extractBucketName(input.Url); bucket != "" {
			buckets[bucket] = true
		}
	}

	for _, output := range task.Outputs {
		if bucket := extractBucketName(output.Url); bucket != "" {
			buckets[bucket] = true
		}
	}

	// Mount all buckets to `/mnt/share/<BUCKET>` as volumes in the GCP Job Request
	var volumes []*batchpb.Volume
	for bucketName := range buckets {
		volumes = append(volumes, &batchpb.Volume{
			Source: &batchpb.Volume_Gcs{
				Gcs: &batchpb.GCS{
					RemotePath: bucketName,
				},
			},
			MountPath: fmt.Sprintf("/mnt/disks/%s", bucketName),
		})
	}

	// Runnables
	var runnables []*batchpb.Runnable
	for _, executor := range task.Executors {
		cmd := strings.Join(executor.Command, " ")

		if executor.Stdout != "" {
			// Redirect command output to the specified file path
			cmd = fmt.Sprintf("%s | tee %s", cmd, executor.Stdout)
		}

		runnable := &batchpb.Runnable{
			Executable: &batchpb.Runnable_Container_{
				Container: &batchpb.Runnable_Container{
					ImageUri: executor.Image,
					Commands: []string{"sh", "-c", cmd},
				},
			},
		}

		runnables = append(runnables, runnable)
	}

	// Resources
	resources := &batchpb.ComputeResource{}
	if task.Resources != nil {
		resources = &batchpb.ComputeResource{
			CpuMilli:  int64(task.Resources.CpuCores) * 1000,
			MemoryMib: int64(task.Resources.RamGb) * 1024,
		}
	}

	// TaskSpec
	taskSpec := &batchpb.TaskSpec{
		Runnables:       runnables,
		ComputeResource: resources,
		Environment:     &batchpb.Environment{},
		Volumes:         volumes,
	}

	// TaskGroup
	taskGroup := &batchpb.TaskGroup{
		TaskCount: 1,
		TaskSpec:  taskSpec,
	}

	// JobRequest
	req := &batchpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", b.conf.Project, b.conf.Location),
		JobId:  task.Id,
		Job: &batchpb.Job{
			Name:       task.Id,
			Uid:        task.Id,
			TaskGroups: []*batchpb.TaskGroup{taskGroup},
			LogsPolicy: &batchpb.LogsPolicy{
				Destination: batchpb.LogsPolicy_CLOUD_LOGGING,
			},
		},
	}

	logger.Debug("GCP Batch Job Request", "request", req)

	resp, err := b.client.CreateJob(context.Background(), req)

	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to GCP Batch",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

	logger.Debug("Submitted task to GCP Batch",
		"taskID", task.Id,
		"gcpbatch_uid", resp.Uid,
		"gcpbatch_name", resp.Name)

	return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{
			"gcpbatch_uid":  resp.Uid,
			"gcpbatch_name": resp.Name,
		}),
	)
}

func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	// TODO: Implement
	return nil
}

// Reconciler adapted from aws_batch/backend.go
//
// Currently the logic is to:
//  1. List all tasks in QUEUED, INITIALIZING, RUNNING states from the Funnel Database
//  2. Map all TES Task IDs to GCP Job IDs
//  3. List all GCP Jobs that have FAILED
//  4. Update the TES Task in the Funnel Database to SYSTEM_ERROR
//
// NOTE: Successful Jobs will be handled
//
// Reconcile loops through tasks and checks the status from Funnel's database
// against the status reported by GCP Batch. This allows the backend to report
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
	ticker := time.NewTicker(b.conf.ReconcileRate.AsDuration())

ReconcileLoop:
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pageToken := ""
			states := []tes.State{tes.Queued, tes.Initializing, tes.Running}

			// For all Task states in QUEUED, INITIALIZING, RUNNING
			for _, s := range states {
				for {
					fmt.Println("DEBUG: s:", s)

					// List Tasks from Funnel Database
					lresp, err := b.database.ListTasks(ctx, &tes.ListTasksRequest{
						View:      tes.View_BASIC.String(),
						State:     s,
						PageSize:  100,
						PageToken: pageToken,
					})
					if err != nil {
						b.log.Error("Calling ListTasks", err)
						continue ReconcileLoop
					}
					pageToken = lresp.NextPageToken

					fmt.Println("DEBUG: lresp:", lresp)

					// Map TES Task ID → GCP Batch Job ID
					tmap := make(map[string]*tes.Task)
					var jobs []*string
					for _, t := range lresp.Tasks {
						jobid := getTaskID(t)
						if jobid != "" {
							tmap[jobid] = t
							jobs = append(jobs, aws.String(jobid))
						}
					}

					// Last page of jobs from the Funnel Database
					if len(jobs) == 0 {
						if pageToken == "" {
							break
						}
						continue
					}

					// List jobs from GCP Batch
					req := &batchpb.ListJobsRequest{
						Parent: fmt.Sprintf("projects/%s/locations/%s", b.conf.Project, b.conf.Location),
					}

					// Using an iterator here to page through GCP Batch jobs
					// Ref: https://pkg.go.dev/cloud.google.com/go/batch@v1.13.0/apiv1#example-Client.ListJobs
					it := b.client.ListJobs(ctx, req)

					for {
						j, err := it.Next()
						if err != nil {
							break
						}

						// If Job is in our list
						if _, ok := tmap[j.Uid]; !ok {
							continue
						}

						fmt.Println("DEBUG: j:", j)

						// task := tmap[*j.JobId]
						// jstate := *j.Status

						// // Failed Jobs
						// if jstate == "FAILED" {
						// 	b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
						// 	b.event.WriteEvent(
						// 		ctx,
						// 		events.NewSystemLog(
						// 			task.Id, 0, 0, "error",
						// 			"GCP Batch job in FAILED state",
						// 			map[string]string{"error": *j.StatusReason, "gcpbatch_id": *j.JobId},
						// 		),
						// 	)
						// }
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
}

// Retreives Batch Job ID from Task metadata (created in #Submit) stored in Funnel Database
func getTaskID(task *tes.Task) string {
	logs := task.GetLogs()
	if len(logs) > 0 {
		metadata := logs[0].GetMetadata()
		if metadata != nil {
			return metadata["gcpbatch_uid"]
		}
	}
	return ""
}
