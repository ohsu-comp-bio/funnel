package gcp_batch

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestCreateJobc(t *testing.T) {
	log := logger.NewLogger("test", logger.DefaultConfig())

	conf := config.DefaultConfig()

	compute, err := NewBackend(context.Background(), conf.GCPBatch, nil, nil, log.Sub("gcp-batch"))
	if err != nil {
		t.Fatal(err)
	}

	task := &tes.Task{
		Id: "task1",
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	err = compute.Submit(task)
	if err != nil {
		t.Fatal(err)
	}
}
