package resources

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Create the Worker/Executor ServiceAccount from config/kubernetes-serviceaccount.yaml
func CreateServiceAccount(ctx context.Context, task *tes.Task, conf *config.Config, client kubernetes.Interface, log *logger.Logger, ownerRef *metav1.OwnerReference) error {

	// Load templates
	t, err := template.New(task.Id).Parse(conf.Kubernetes.ServiceAccountTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	// TODO: Handle cases where values/tags below are not supplied
	var buf bytes.Buffer
	templateData := map[string]interface{}{
		"TaskId":             task.Id,
		"Namespace":          conf.Kubernetes.JobsNamespace,
		"ServiceAccountName": fmt.Sprintf("funnel-worker-sa-%s-%s", conf.Kubernetes.JobsNamespace, task.Id),
	}

	// Override ServiceAccountName if provided in Task Tags
	if saName, exists := task.Tags["_WORKER_SA"]; exists && saName != "" {
		templateData["ServiceAccountName"] = saName
	}

	// Only include IamRoleArn if provided
	if roleArn, exists := task.Tags["_FUNNEL_WORKER_ROLE_ARN"]; exists && roleArn != "" {
		templateData["IamRoleArn"] = roleArn
	}

	err = t.Execute(&buf, templateData)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to decode ServiceAccount spec: %v", err)
	}

	sa, ok := obj.(*corev1.ServiceAccount)
	if !ok {
		return fmt.Errorf("failed to cast to ServiceAccount spec")
	}

	if ownerRef != nil {
		sa.OwnerReferences = []metav1.OwnerReference{*ownerRef}
	}

	_, err = client.CoreV1().ServiceAccounts(conf.Kubernetes.JobsNamespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ServiceAccount: %v", err)
	}

	return nil
}

func isServiceAccountAttachedToPods(ctx context.Context, saName, namespace string, client kubernetes.Interface) (bool, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.serviceAccountName=%s", saName),
	})
	if err != nil {
		return false, fmt.Errorf("listing pods using ServiceAccount %s: %v", saName, err)
	}
	return len(pods.Items) > 0, nil
}

// DeleteServiceAccount deletes the ServiceAccount created for a task.
// If externalSA is true the ServiceAccount is externally managed (e.g. a
// Gen3Workflow per-user SA supplied via the _WORKER_SA task tag) and must not
// be deleted by Funnel.
func DeleteServiceAccount(ctx context.Context, taskID string, namespace string, client kubernetes.Interface, log *logger.Logger, externalSA bool) error {
	if externalSA {
		log.Debug("skipping deletion of externally-managed ServiceAccount", "taskID", taskID)
		return nil
	}
	// ServiceAccount names are not available here without config, so we list by label.
	sas, err := client.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=funnel,taskId=%s", taskID),
	})
	if err != nil {
		return fmt.Errorf("listing ServiceAccounts for task %s: %v", taskID, err)
	}
	for _, sa := range sas.Items {
		inUse, err := isServiceAccountAttachedToPods(ctx, sa.Name, namespace, client)
		if err != nil {
			return err
		}
		if inUse {
			return fmt.Errorf("serviceAccount %s is still in use by active pod(s)", sa.Name)
		}

		log.Debug("deleting Worker ServiceAccount", "name", sa.Name, "taskID", taskID)
		if err := client.CoreV1().ServiceAccounts(namespace).Delete(ctx, sa.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("deleting ServiceAccount %s: %v", sa.Name, err)
		}
	}
	return nil
}
