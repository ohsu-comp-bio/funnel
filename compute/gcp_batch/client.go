package gcp_batch

import (
	"context"

	batch "cloud.google.com/go/batch/apiv1"
	"cloud.google.com/go/batch/apiv1/batchpb"
	"github.com/googleapis/gax-go/v2"
)

// client defines the subset of batch.Client methods our backend uses.
// This makes it easy to mock for testing.
type client interface {
	CreateJob(ctx context.Context, req *batchpb.CreateJobRequest, opts ...gax.CallOption) (*batchpb.Job, error)
}

// Ensure the real client satisfies our interface.
var _ client = (*batch.Client)(nil)
