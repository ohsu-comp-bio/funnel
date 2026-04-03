// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

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
  name: funnel-{{.TaskId}}
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
	conf.Kubernetes.ServiceAccountTemplate = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.ServiceAccountName}}
  namespace: {{.Namespace}}
`
	conf.Kubernetes.RoleTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: funnel-worker-role-{{.TaskId}}
  namespace: {{.Namespace}}
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
`
	conf.Kubernetes.RoleBindingTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: funnel-worker-rolebinding-{{.TaskId}}
  namespace: {{.Namespace}}
subjects:
- kind: ServiceAccount
  name: {{.ServiceAccountName}}
  namespace: {{.Namespace}}
roleRef:
  kind: Role
  name: funnel-worker-role-{{.TaskId}}
  apiGroup: rbac.authorization.k8s.io
`

	// Create a logger
	log := logger.NewLogger("test", logger.DefaultConfig())

	backend := &Backend{
		client:   fakeClient,
		event:    nil,
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
	job, err := fakeClient.BatchV1().Jobs(conf.Kubernetes.Namespace).Get(context.Background(), "funnel-"+task.Id, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	if job.Name != "funnel-"+task.Id {
		t.Errorf("expected Job name 'funnel-%s', got '%s'", task.Id, job.Name)
	}

	// Verify that the ConfigMap was created
	configMapName := "funnel-worker-config-" + task.Id
	_, err = fakeClient.CoreV1().ConfigMaps(conf.Kubernetes.Namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get ConfigMap: %v", err)
	}

	// Clean up resources
	err = backend.cleanResources(context.Background(), task.Id)
	if err != nil {
		t.Fatalf("failed to clean resources: %v", err)
	}

	// Verify that the Job was deleted
	_, err = fakeClient.BatchV1().Jobs(conf.Kubernetes.Namespace).Get(context.Background(), "funnel-"+task.Id, metav1.GetOptions{})
	if err == nil {
		t.Error("expected Job to be deleted, but it still exists")
	}

	// Verify that the ConfigMap was deleted
	_, err = fakeClient.CoreV1().ConfigMaps(conf.Kubernetes.Namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err == nil {
		t.Error("expected ConfigMap to be deleted, but it still exists")
	}
}
