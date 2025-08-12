// Package gcp_batch contains code for accessing compute resources via Google Batch.
// ref: https://cloud.google.com/batch/docs
package gcp_batch

import (
	"context"

	batch "cloud.google.com/go/batch/apiv1"
	"cloud.google.com/go/batch/apiv1/batchpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type Backend struct {
	client   *batch.Client
	conf     config.GCPBatch
	event    events.Writer
	database tes.ReadOnlyServer
	log      *logger.Logger
	events.Computer
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
	ctx := context.Background()

	// TODO: Implement proper Google Batch job creation
	// This is a placeholder implementation that needs to be completed
	// with proper Google Batch API structures

	// TODO: The GoogleBatch config needs a Parent field for the project/location
	// For now using a placeholder value
	parent := "projects/my-project/locations/us-central1" // TODO: get from config

	req := &batchpb.CreateJobRequest{
		Parent: parent,
		JobId:  task.Id,
		Job:    &batchpb.Job{
			// TODO: Configure the job properly based on the TES task
			// This would include:
			// - TaskGroups with TaskSpecs containing Runnables
			// - AllocationPolicy for compute resources
			// - LogsPolicy for output handling
		},
	}

	resp, err := b.client.CreateJob(ctx, req)
	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to Google Batch",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

	return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{"gcp_batch_id": resp.GetName()}),
	)
}

func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	// TODO: Implement this
	return nil
}

func (b *Backend) reconcile(ctx context.Context) {
	// TODO: Implement this
}
