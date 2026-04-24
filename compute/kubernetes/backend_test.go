// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"context"
	"reflect"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// noopEventWriter implements events.Writer for testing.
type noopEventWriter struct{}

func (n *noopEventWriter) WriteEvent(ctx context.Context, ev *events.Event) error {
	return nil
}

func (n *noopEventWriter) Close() {}

func TestTaskSubmission(t *testing.T) {
	// Create a fake Kubernetes client
	fakeClient := fake.NewSimpleClientset()

	// Create a mock configuration
	conf := config.DefaultConfig()
	conf.Kubernetes.Namespace = "test-namespace"
	conf.Kubernetes.JobsNamespace = "test-namespace"
	conf.Kubernetes.WorkerTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.TaskId}}
  namespace: {{.JobsNamespace}}
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: task
        image: alpine
        command: ["echo", "hello world"]
        resources:
          requests:
            cpu: "{{.Cpus}}"
            memory: "{{.RamGb}}Gi"
            ephemeral-storage: "{{.DiskGb}}Gi"
`
	conf.Kubernetes.ConfigMapTemplate = `apiVersion: v1
kind: ConfigMap
metadata:
  name: funnel-worker-config-{{.TaskId}}
  namespace: {{.Namespace}}
data:
  config.yaml: "placeholder"
`

	// Create a logger
	log := logger.NewLogger("test", logger.DefaultConfig())

	backend := &Backend{
		client:   fakeClient,
		event:    &noopEventWriter{},
		database: nil,
		log:      log,
		conf:     conf, // Funnel configuration
	}

	// Define a test task
	task := &tes.Task{
		Id: "test-task",
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1.0,
			DiskGb:   10.0,
		},
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	// Submit the task to the backend
	err := backend.Submit(context.Background(), task, conf)
	if err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	// Verify that the Job was created
	job, err := fakeClient.BatchV1().Jobs(conf.Kubernetes.JobsNamespace).Get(context.Background(), task.Id, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	if job.Name != task.Id {
		t.Errorf("expected Job name '%s', got '%s'", task.Id, job.Name)
	}

	// Verify that the ConfigMap was created
	configMapName := "funnel-worker-config-" + task.Id
	_, err = fakeClient.CoreV1().ConfigMaps(conf.Kubernetes.JobsNamespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get ConfigMap: %v", err)
	}

	// Clean up resources
	err = backend.cleanResources(context.Background(), task.Id)
	if err != nil {
		t.Fatalf("failed to clean resources: %v", err)
	}

	// Verify that the Job was deleted
	_, err = fakeClient.BatchV1().Jobs(conf.Kubernetes.JobsNamespace).Get(context.Background(), task.Id, metav1.GetOptions{})
	if err == nil {
		t.Error("expected Job to be deleted, but it still exists")
	}

	// Verify that the ConfigMap was deleted
	_, err = fakeClient.CoreV1().ConfigMaps(conf.Kubernetes.JobsNamespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err == nil {
		t.Error("expected ConfigMap to be deleted, but it still exists")
	}

}

func TestSubmit_AppliesNodeSelectorAndTolerationsToWorkerJob(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a fake funnel server pod so CreateJob can resolve worker image.
	funnelPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-server",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "funnel",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "funnel", Image: "alpine"},
			},
		},
	}
	if _, err := fakeClient.CoreV1().Pods("test-namespace").Create(context.Background(), funnelPod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("failed to create fake funnel pod: %v", err)
	}

	conf := config.DefaultConfig()
	conf.Compute = "kubernetes"
	conf.Kubernetes.Namespace = "test-namespace"
	conf.Kubernetes.JobsNamespace = "test-namespace"
	conf.Kubernetes.NodeSelector = map[string]string{
		"pool": "cpu",
		"zone": "us-west-2a",
	}
	conf.Kubernetes.Tolerations = []*config.Toleration{
		{
			Key:      "dedicated",
			Operator: "Equal",
			Value:    "worker",
			Effect:   "NoSchedule",
		},
	}

	// Include scheduling blocks in the worker template under test.
	conf.Kubernetes.WorkerTemplate = `
apiVersion: batch/v1
kind: Job
metadata:
  name: funnel-{{.TaskId}}
  namespace: {{.JobsNamespace}}
spec:
  template:
    spec:
      {{- if .NodeSelector }}
      nodeSelector:
        {{- range $key, $value := .NodeSelector }}
        {{ $key }}: {{ $value }}
        {{- end }}
      {{- end }}
      {{- if .Tolerations }}
      tolerations:
        {{- range .Tolerations }}
        - key: {{ .Key }}
          operator: {{ .Operator }}
          effect: {{ .Effect }}
          {{if .Value}}value: {{ .Value }}{{end}}
        {{- end }}
      {{- end }}
      restartPolicy: Never
      containers:
      - name: worker
        image: alpine
        command: ["echo", "ok"]
`

	log := logger.NewLogger("test", logger.DefaultConfig())
	backend := &Backend{
		client: fakeClient,
		log:    log,
		conf:   conf,
		event:  &noopEventWriter{},
	}

	task := &tes.Task{
		Id: "sched-worker",
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1,
			DiskGb:   1,
		},
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "ok"},
			},
		},
	}

	if err := backend.Submit(context.Background(), task, conf); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	job, err := fakeClient.BatchV1().Jobs(conf.Kubernetes.JobsNamespace).Get(context.Background(), "funnel-"+task.Id, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get worker job: %v", err)
	}

	if !reflect.DeepEqual(job.Spec.Template.Spec.NodeSelector, conf.Kubernetes.NodeSelector) {
		t.Fatalf("nodeSelector mismatch: got=%v want=%v", job.Spec.Template.Spec.NodeSelector, conf.Kubernetes.NodeSelector)
	}

	if len(job.Spec.Template.Spec.Tolerations) != 1 {
		t.Fatalf("unexpected tolerations length: got=%d want=1", len(job.Spec.Template.Spec.Tolerations))
	}

	gotTol := job.Spec.Template.Spec.Tolerations[0]
	if gotTol.Key != "dedicated" {
		t.Fatalf("unexpected toleration key: got=%q want=%q", gotTol.Key, "dedicated")
	}
	if gotTol.Operator != corev1.TolerationOperator("Equal") {
		t.Fatalf("unexpected toleration operator: got=%q want=%q", gotTol.Operator, corev1.TolerationOperator("Equal"))
	}
	if gotTol.Effect != corev1.TaintEffect("NoSchedule") {
		t.Fatalf("unexpected toleration effect: got=%q want=%q", gotTol.Effect, corev1.TaintEffect("NoSchedule"))
	}
	if gotTol.Value != "worker" {
		t.Fatalf("unexpected toleration value: got=%q want=%q", gotTol.Value, "worker")
	}
}
