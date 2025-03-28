package kubernetes

import (
	"fmt"
	"os"
	"testing"

	"github.com/ohsu-comp-bio/funnel/compute/kubernetes/resources"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestCreateJob(t *testing.T) {
	conf := config.DefaultConfig().Kubernetes

	content, err := os.ReadFile("../../config/kubernetes-template.yaml")
	if err != nil {
		t.Fatal(fmt.Errorf("reading template: %v", err))
	}

	conf.Template = string(content)

	log := logger.NewLogger("test", logger.DefaultConfig())

	b := &Backend{
		client:    nil,
		namespace: conf.Namespace,
		template:  conf.Template,
		event:     nil,
		database:  nil,
		log:       log,
	}

	t.Run("SuccessfulJobCreation", func(t *testing.T) {
		task := &tes.Task{
			Id: "task1",
			Executors: []*tes.Executor{
				{
					Image:   "alpine",
					Command: []string{"echo", "hello world"},
				},
			},
		}

		err := resources.CreateJob(task, b.namespace, b.template)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("MissingTemplate", func(t *testing.T) {
		b.template = ""
		task := &tes.Task{
			Id: "task2",
			Executors: []*tes.Executor{
				{
					Image:   "alpine",
					Command: []string{"echo", "hello world"},
				},
			},
		}

		err := resources.CreateJob(task, b.namespace, b.template)
		if err == nil {
			t.Fatal("expected error for missing template, got none")
		}
	})

	t.Run("InvalidTaskDefinition", func(t *testing.T) {
		task := &tes.Task{
			Id:        "task3",
			Executors: []*tes.Executor{}, // No executors defined
		}

		err := resources.CreateJob(task, b.namespace, b.template)
		if err == nil {
			t.Fatal("expected error for invalid task definition, got none")
		}
	})
}
