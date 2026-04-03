package resources

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace     = "test-namespace"
	jobsNamespace = "test-jobs-namespace"
	testTaskID    = "test-task-id"
)

var l = logger.NewLogger("test", logger.DefaultConfig())

// newTestConfig returns a Config with Kubernetes settings populated for testing.
func newTestConfig() *config.Config {
	return &config.Config{
		Kubernetes: &config.Kubernetes{
			Namespace:     namespace,
			JobsNamespace: jobsNamespace,
			WorkerTemplate: `apiVersion: batch/v1
kind: Job
metadata:
  name: funnel-{{.TaskId}}
  namespace: {{.JobsNamespace}}
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: funnel-worker
        image: ohsucompbio/funnel:latest
        command: ["funnel", "worker", "run", "--task-id", "{{.TaskId}}"]
`,
			PVTemplate: `apiVersion: v1
kind: PersistentVolume
metadata:
  name: funnel-worker-pv-{{.TaskId}}
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  storageClassName: manual
  hostPath:
    path: /tmp/funnel-worker-{{.TaskId}}
`,
			PVCTemplate: `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: funnel-worker-pvc-{{.TaskId}}
  namespace: {{.Namespace}}
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: manual
  resources:
    requests:
      storage: 10Gi
`,
			ServiceAccountTemplate: `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.ServiceAccountName}}
  namespace: {{.Namespace}}
`,
			RoleTemplate: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: funnel-worker-role-{{.TaskId}}
  namespace: {{.Namespace}}
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
`,
			RoleBindingTemplate: `apiVersion: rbac.authorization.k8s.io/v1
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
`,
		},
		GenericS3: []*config.GenericS3Storage{
			{Bucket: "test-bucket", Region: "us-west-2"},
		},
	}
}

func TestCreateConfigMap(t *testing.T) {
	conf := &config.Config{
		Kubernetes: &config.Kubernetes{
			JobsNamespace: jobsNamespace,
		},
	}
	err := CreateConfigMap(testTaskID, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateConfigMap failed: %v", err)
	}
}

func TestDeleteConfigMap(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a test ConfigMap first
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-worker-config-" + testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test ConfigMap: %v", err)
	}

	err = DeleteConfigMap(context.Background(), testTaskID, namespace, fakeClient, l)
	if err != nil {
		t.Errorf("DeleteConfigMap failed: %v", err)
	}

	// Verify deletion
	_, err = fakeClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), "funnel-worker-"+testTaskID, metav1.GetOptions{})
	if err == nil {
		t.Error("ConfigMap was not deleted")
	}
}

func TestCreateJob(t *testing.T) {
	task := &tes.Task{
		Id: testTaskID,
		Resources: &tes.Resources{
			CpuCores: 1,
			RamGb:    1.0,
			DiskGb:   10.0,
		},
	}

	conf := newTestConfig()
	err := CreateJob(task, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateJob failed: %v", err)
	}
}

func TestDeleteJob(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a test Job first (name must match funnel-{taskID} prefix used by CreateJob/DeleteJob)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-" + testTaskID,
			Namespace: jobsNamespace,
		},
	}
	_, err := fakeClient.BatchV1().Jobs(jobsNamespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test Job: %v", err)
	}

	conf := newTestConfig()
	err = DeleteJob(context.Background(), conf, testTaskID, fakeClient, l)
	if err != nil {
		t.Errorf("DeleteJob failed: %v", err)
	}

	// Verify deletion
	_, err = fakeClient.BatchV1().Jobs(jobsNamespace).Get(context.Background(), "funnel-"+testTaskID, metav1.GetOptions{})
	if err == nil {
		t.Error("Job was not deleted")
	}
}

func TestCreatePV(t *testing.T) {
	conf := newTestConfig()
	err := CreatePV(testTaskID, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreatePV failed: %v", err)
	}
}

func TestDeletePV(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a test PV first
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "funnel-worker-pv-" + testTaskID,
		},
	}
	_, err := fakeClient.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test PV: %v", err)
	}

	err = DeletePV(context.Background(), testTaskID, fakeClient, l)
	if err != nil {
		t.Errorf("DeletePV failed: %v", err)
	}

	// Verify deletion
	_, err = fakeClient.CoreV1().PersistentVolumes().Get(context.Background(), "funnel-worker-pv-"+testTaskID, metav1.GetOptions{})
	if err == nil {
		t.Error("PV was not deleted")
	}
}

func TestCreatePVC(t *testing.T) {
	conf := newTestConfig()
	err := CreatePVC(testTaskID, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreatePVC failed: %v", err)
	}
}

func TestDeletePVC(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a test PVC first
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-worker-pvc-" + testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test PVC: %v", err)
	}

	err = DeletePVC(context.Background(), testTaskID, namespace, fakeClient, l)
	if err != nil {
		t.Errorf("DeletePVC failed: %v", err)
	}

	// Verify deletion
	_, err = fakeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), "funnel-worker-pvc-"+testTaskID, metav1.GetOptions{})
	if err == nil {
		t.Error("PVC was not deleted")
	}
}

func TestCreateJobWithNoResources(t *testing.T) {
	task := &tes.Task{
		Id: testTaskID,
		// Intentionally omit Resources to test default handling
	}

	conf := newTestConfig()
	err := CreateJob(task, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateJob failed with nil resources: %v", err)
	}
}

func TestDeleteNonExistentResources(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	nonExistentID := "non-existent-id"

	// Test deleting non-existent resources
	t.Run("ConfigMap", func(t *testing.T) {
		err := DeleteConfigMap(context.Background(), nonExistentID, namespace, fakeClient, l)
		if err == nil {
			t.Error("Expected error when deleting non-existent ConfigMap")
		}
	})

	// DeletePV and DeletePVC are idempotent - they don't error on non-existent resources
	// because PV/PVC may never have been created for tasks without I/O.
	t.Run("PV", func(t *testing.T) {
		err := DeletePV(context.Background(), nonExistentID, fakeClient, l)
		if err != nil {
			t.Errorf("Unexpected error when deleting non-existent PV: %v", err)
		}
	})

	t.Run("PVC", func(t *testing.T) {
		err := DeletePVC(context.Background(), nonExistentID, namespace, fakeClient, l)
		if err != nil {
			t.Errorf("Unexpected error when deleting non-existent PVC: %v", err)
		}
	})
}

func TestCreateServiceAccount(t *testing.T) {
	task := &tes.Task{
		Id: testTaskID,
		Tags: map[string]string{
			"funnel_worker_role_arn": "arn:aws:iam::123456789012:role/funnel-worker-role",
		},
	}

	conf := newTestConfig()
	err := CreateServiceAccount(task, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateServiceAccount failed: %v", err)
	}
}

func TestDeleteServiceAccount(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-worker-sa-" + testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.CoreV1().ServiceAccounts(namespace).Create(context.Background(), sa, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test ServiceAccount: %v", err)
	}

	err = DeleteServiceAccount(context.Background(), testTaskID, fakeClient, l)
	if err != nil {
		t.Errorf("DeleteServiceAccount failed: %v", err)
	}
}

func TestCreateRole(t *testing.T) {
	task := &tes.Task{
		Id: testTaskID,
	}

	conf := newTestConfig()
	err := CreateRole(task, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateRole failed: %v", err)
	}
}

func TestDeleteRole(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	role := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-worker-role-" + testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.CoreV1().ServiceAccounts(namespace).Create(context.Background(), role, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test Role: %v", err)
	}

	err = DeleteRole(context.Background(), testTaskID, fakeClient, l)
	if err != nil {
		t.Errorf("DeleteRole failed: %v", err)
	}
}

func TestCreateRoleBinding(t *testing.T) {
	task := &tes.Task{
		Id: testTaskID,
	}

	conf := newTestConfig()
	err := CreateRoleBinding(task, conf, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateRoleBinding failed: %v", err)
	}
}

func TestDeleteRoleBinding(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	roleBinding := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "funnel-worker-rolebinding-" + testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.CoreV1().ServiceAccounts(namespace).Create(context.Background(), roleBinding, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test RoleBinding: %v", err)
	}

	err = DeleteRoleBinding(context.Background(), testTaskID, fakeClient, l)
	if err != nil {
		t.Errorf("DeleteRoleBinding failed: %v", err)
	}
}
