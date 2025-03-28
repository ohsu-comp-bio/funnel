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

// Create the Worker/Executor PV from config/kubernetes-pv.yaml
// TODO: Move this config file to Helm Charts so users can see/customize it
func CreatePV(taskId string, namespace string, bucket string, region string, tpl string) error {
	// Load templates
	t, err := template.New(taskId).Parse(tpl)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"Namespace": namespace,
		"Bucket":    bucket,
		"Region":    region,
	})
	if err != nil {
		return fmt.Errorf("executing PV template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("decoding PV spec: %v", err)
	}

	_, ok := obj.(*corev1.PersistentVolume)
	if !ok {
		return fmt.Errorf("failed to decode PV spec")
	}
	return nil
}

// Add this helper function for PV cleanup
func DeletePV(ctx context.Context, taskID string, client kubernetes.Interface) error {
	name := fmt.Sprintf("funnel-pv-%s", taskID)
	err := client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("deleting shared PV: %v", err)
	}

	return nil
}
