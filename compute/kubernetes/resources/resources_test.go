package resources

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/plugins/proto"
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

func TestCreateConfigMap(t *testing.T) {
	conf := &config.Config{}
	err := CreateConfigMap(testTaskID, namespace, conf, fake.NewSimpleClientset(), l)
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

	tpl := `
apiVersion: batch/v1
kind: Job
metadata:
  name: funnel-{{.TaskId}}
  namespace: {{.Namespace}}
spec:
  template:
    spec:
      containers:
      - name: task
        resources:
          requests:
            cpu: "{{.Cpus}}"
            memory: "{{.RamGb}}Gi"
            ephemeral-storage: "{{.DiskGb}}Gi"
`

	err := CreateJob(task, namespace, jobsNamespace, tpl, fake.NewSimpleClientset(), l)
	if err != nil {
		t.Errorf("CreateJob failed: %v", err)
	}
}

func TestDeleteJob(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Create a test Job first
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testTaskID,
			Namespace: namespace,
		},
	}
	_, err := fakeClient.BatchV1().Jobs(namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test Job: %v", err)
	}

	err = DeleteJob(context.Background(), testTaskID, fakeClient.BatchV1().Jobs(namespace), l)
	if err != nil {
		t.Errorf("DeleteJob failed: %v", err)
	}

	// Verify deletion
	_, err = fakeClient.BatchV1().Jobs(namespace).Get(context.Background(), testTaskID, metav1.GetOptions{})
	if err == nil {
		t.Error("Job was not deleted")
	}
}

func TestCreatePV(t *testing.T) {
	tpl := `
apiVersion: v1
kind: PersistentVolume
metadata:
  name: funnel-worker-pv-{{.TaskId}}
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  awsElasticBlockStore:
    volumeID: {{.Bucket}}
    fsType: ext4
`

	err := CreatePV(testTaskID, namespace, "test-bucket", "test-region", tpl, fake.NewSimpleClientset(), l)
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
	tpl := `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: funnel-worker-pvc-{{.TaskId}}
  namespace: {{.Namespace}}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`

	err := CreatePVC(testTaskID, namespace, "test-bucket", "test-region", tpl, fake.NewSimpleClientset(), l)
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

	tpl := `
apiVersion: batch/v1
kind: Job
metadata:
  name: funnel-{{.TaskId}}
  namespace: {{.Namespace}}
spec:
  template:
    spec:
      containers:
      - name: task
        resources:
          requests:
            cpu: "{{.Cpus}}"
            memory: "{{.RamGb}}Gi"
            ephemeral-storage: "{{.DiskGb}}Gi"
`

	err := CreateJob(task, namespace, jobsNamespace, tpl, fake.NewSimpleClientset(), l)
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

	t.Run("PV", func(t *testing.T) {
		err := DeletePV(context.Background(), nonExistentID, fakeClient, l)
		if err == nil {
			t.Error("Expected error when deleting non-existent PV")
		}
	})

	t.Run("PVC", func(t *testing.T) {
		err := DeletePVC(context.Background(), nonExistentID, namespace, fakeClient, l)
		if err == nil {
			t.Error("Expected error when deleting non-existent PVC")
		}
	})
}

func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()
	dst := &config.Config{
		Kubernetes: &config.Kubernetes{
			Namespace: "original-namespace",
		},
	}

	// Test with nil response in context
	t.Run("NilResponse", func(t *testing.T) {
		err := UpdateConfig(ctx, dst)
		if err == nil {
			t.Error("Expected error with nil response in context")
		}
	})

	// Test with valid config merge
	t.Run("ValidMerge", func(t *testing.T) {
		pluginResp := &proto.GetResponse{
			Config: &config.Config{
				Kubernetes: &config.Kubernetes{
					Namespace: "new-namespace",
				},
			},
		}
		ctxWithResp := context.WithValue(ctx, "pluginResponse", pluginResp)

		err := UpdateConfig(ctxWithResp, dst)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if dst.Kubernetes.Namespace != "new-namespace" {
			t.Errorf("Expected namespace to be updated to 'new-namespace', got %s", dst.Kubernetes.Namespace)
		}
	})
}
