package worker

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/tes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var taskTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.TaskId}}-{{.JobId}}
spec:
  template:
	spec:
	  serviceAccountName: {{.ServiceAccount}}
	  containers:
	  - name: {{.TaskId}}
		image: {{.Image}}
		command: {{.Command}}
		resources:
		  requests:
			cpu: "{{.Cpus}}"
			memory: "{{.RamGb}}Gi"
			ephemeral-storage: "{{.DiskGb}}Gi"
	  restartPolicy: Never
`

func TestKubernetesRun(t *testing.T) {
	kcmd := KubernetesCommand{
		TaskId:       "test-task",
		JobId:        1,
		TaskTemplate: taskTemplate,
		Namespace:    "default",
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1,
			DiskGb:   1,
		},
		Command: Command{
			Image:        "alpine",
			ShellCommand: []string{"echo", "Hello, World!"},
		},
	}

	clientset := fake.NewSimpleClientset()

	ctx := context.Background()
	err := kcmd.Run(ctx)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	jobs, err := clientset.BatchV1().Jobs(kcmd.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if len(jobs.Items) != 1 {
		t.Errorf("Expected 1 job, but got: %d", len(jobs.Items))
	}
}
