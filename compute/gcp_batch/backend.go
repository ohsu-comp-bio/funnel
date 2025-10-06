// Package gcp_batch contains code for accessing compute resources via Google Batch.
// ref: https://cloud.google.com/batch/docs
// ref: https://cloud.google.com/batch/docs/reference/rest
package gcp_batch

import (
	"context"
	"encoding/json"
	"fmt"

	batch "cloud.google.com/go/batch/apiv1"
	"cloud.google.com/go/batch/apiv1/batchpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type Backend struct {
	client   client
	conf     config.GCPBatch
	event    events.Writer
	database tes.ReadOnlyServer
	log      *logger.Logger
	events.Backend
}

func NewBackend(ctx context.Context, conf config.GCPBatch, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
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

func (b *Backend) Submit(task *tes.Task) error {

	// Pretty print the TES Task for debugging
	if taskJSON, err := json.MarshalIndent(task, "", "  "); err == nil {
		fmt.Printf("TES Task:\n%s\n", string(taskJSON))
	} else {
		fmt.Printf("TES Task: %+v\n", task)
	}

	b.conf.Project = "tes-batch-integration-test" // TODO: Remove hardcoding
	b.conf.Location = "us-central1"               // TODO: Remove hardcoding

	fmt.Printf("DEBUG: b.conf.Project: %s\n", b.conf.Project)
	fmt.Printf("DEBUG: b.conf.Location: %s\n", b.conf.Location)

	runnable := &batchpb.Runnable{
		Executable: &batchpb.Runnable_Script_{
			Script: &batchpb.Runnable_Script{
				Command: &batchpb.Runnable_Script_Text{
					Text: "echo Hello, world!",
				},
			},
		},
	}

	// 2. Create the TaskSpec with the runnable(s)
	taskSpec := &batchpb.TaskSpec{
		Runnables: []*batchpb.Runnable{runnable},
		// Add compute requirements, environment variables, etc., if needed
	}

	// Create a minimal TaskGroup to satisfy GCP Batch API requirements
	taskGroup := &batchpb.TaskGroup{
		TaskCount: 1,
		TaskSpec:  taskSpec,
	}

	req := &batchpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", b.conf.Project, b.conf.Location),
		JobId:  task.Id,
		Job: &batchpb.Job{
			Name:       task.Id,
			Uid:        task.Id,
			TaskGroups: []*batchpb.TaskGroup{taskGroup},
		},
	}

	// DEBUG: Print the GCP Batch Job request for debugging
	if reqJSON, err := json.MarshalIndent(req, "", "  "); err == nil {
		fmt.Printf("GCP Batch Job Request:\n%s\n", string(reqJSON))
	} else {
		fmt.Printf("GCP Batch Job Request: %+v\n", req)
	}

	// Uncomment to submit the Job to GCP Batch
	fmt.Println("Submitting GCP Batch Job...")
	_, err := b.client.CreateJob(context.Background(), req)
	if err != nil {
		fmt.Printf("Error submitting GCP Batch Job: %v\n", err)
		return err
	}
	fmt.Println("GCP Batch Job submitted successfully.")

	return nil
}

func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	// TODO: Implement this
	return nil
}

func (b *Backend) reconcile(ctx context.Context) {
	// TODO: Implement this
}
