package gcp_batch

import (
	"context"
	"testing"

	"cloud.google.com/go/batch/apiv1/batchpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestCreateJobc(t *testing.T) {
	log := logger.NewLogger("test", logger.DefaultConfig())
	conf := &config.GCPBatch{
		Project:  "example-project",
		Location: "us-west1",
	}

	// Mock client
	// mockClient := &mockClient{
	// 	CreateJobFunc: func(req *batchpb.CreateJobRequest) (*batchpb.Job, error) {
	// 		return &batchpb.Job{Name: "projects/example/locations/us-west1/jobs/example"}, nil
	// 	},
	// }

	// Backend
	compute := &Backend{
		// client: mockClient,
		conf:  conf,
		log:   log,
		event: nil,
	}

	// Task
	task := &tes.Task{
		Id: "task1",
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	err := compute.Submit(task)
	if err != nil {
		t.Fatal(err)
	}
}

type mockClient struct {
	CreateJobFunc func(req *batchpb.CreateJobRequest) (*batchpb.Job, error)
}

func (m *mockClient) CreateJob(ctx context.Context, req *batchpb.CreateJobRequest, opts ...gax.CallOption) (*batchpb.Job, error) {
	if m.CreateJobFunc != nil {
		return m.CreateJobFunc(req)
	}

	return &batchpb.Job{Name: "projects/example/locations/us-west1/jobs/example"}, nil
}
