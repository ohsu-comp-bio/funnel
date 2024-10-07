package kubernetes

import (
	"fmt"
	"os"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

func TestCreateJobc(t *testing.T) {
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

	task := &tes.Task{
		Id: "task1",
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	job, err := b.createJob(task)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", job)
}
