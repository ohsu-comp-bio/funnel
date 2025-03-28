package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Create the Worker/Executor PVC from config/kubernetes-pvc.yaml
// TODO: Move this config file to Helm Charts so users can see/customize it
func CreatePVC(taskId string, namespace string, bucket string, region string, tpl string) error {
	// Load templates
	t, err := template.New(taskId).Parse(tpl)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"TaskId":    taskId,
		"Namespace": namespace,
		"Bucket":    bucket,
		"Region":    region,
	})
	if err != nil {
		return fmt.Errorf("executing PVC template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("decoding PVC spec: %v", err)
	}

	_, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("failed to decode PVC spec")
	}
	return nil
}

// Add this helper function for PVC cleanup
func DeletePVC(ctx context.Context, taskID string, namespace string, client kubernetes.Interface) error {
	name := fmt.Sprintf("funnel-pvc-%s", taskID)
	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("deleting shared PVC: %v", err)
	}

	return nil
}
