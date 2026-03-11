package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Create the Worker/Executor PVC from config/kubernetes-pvc.yaml
// TODO: Move this config file to Helm Charts so users can see/customize it
func CreatePVC(taskId string, config *config.Config, client kubernetes.Interface, log *logger.Logger) error {

	jobNamespace := config.Kubernetes.JobsNamespace

	// Load templates
	t, err := template.New(taskId).Parse(config.Kubernetes.PVCTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	// GenericS3 is optional; only S3 CSI PVCTemplates reference these fields.
	// Deployments using hostPath or other non-S3 PVCTemplates leave them empty.
	s3Bucket, s3Region := "", ""
	if len(config.GenericS3) > 0 {
		s3Bucket = config.GenericS3[0].Bucket
		s3Region = config.GenericS3[0].Region
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"TaskId":    taskId,
		"Namespace": jobNamespace,
		"Bucket":    s3Bucket,
		"Region":    s3Region,
	})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("failed to decode PVC spec")
	}

	_, err = client.CoreV1().PersistentVolumeClaims(jobNamespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// Add this helper function for PVC cleanup
func DeletePVC(ctx context.Context, taskID string, namespace string, client kubernetes.Interface, log *logger.Logger) error {
	name := fmt.Sprintf("funnel-worker-pvc-%s", taskID)
	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		log.Debug("deleting Worker PVC", "taskID", taskID)
		err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deleting shared PVC: %v", err)
		}
	}

	return nil
}
