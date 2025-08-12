// Package gcp_batch contains code for accessing compute resources via Google Batch.
// ref: https://cloud.google.com/batch/docs
// ref: https://pkg.go.dev/cloud.google.com/go/batch/apiv1#hdr-Using_the_Client
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

	req := &batchpb.CreateJobRequest{
		Parent: "projects/my-project/locations/us-west1", // TODO: get from config
		JobId:  task.Id,
		Job: &batchpb.Job{
			Name: task.Id,
			Uid:  task.Id,
		},
	}

	_, err := b.client.CreateJob(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	// TODO: Implement this
	return nil
}

func (b *Backend) reconcile(ctx context.Context) {
	// TODO: Implement this
}
