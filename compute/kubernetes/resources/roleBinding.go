package resources

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Create the Worker/Executor RoleBinding from config/kubernetes-rolebinding.yaml
func CreateRoleBinding(ctx context.Context, task *tes.Task, config *config.Config, client kubernetes.Interface, log *logger.Logger) error {

	// Load templates
	t, err := template.New(task.Id).Parse(config.Kubernetes.RoleBindingTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	// TODO: Handle cases where values/tags below are not supplied
	templateData := map[string]interface{}{
		"TaskId":             task.Id,
		"Namespace":          config.Kubernetes.JobsNamespace,
		"ServiceAccountName": fmt.Sprintf("funnel-worker-sa-%s-%s", config.Kubernetes.JobsNamespace, task.Id),
	}

	// Override ServiceAccountName if provided in Task Tags
	if saName, exists := task.Tags["_WORKER_SA"]; exists && saName != "" {
		templateData["ServiceAccountName"] = saName
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, templateData)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to decode Role spec: %v", err)
	}

	roleBinding, ok := obj.(*rbacv1.RoleBinding) // Change from corev1.Role to rbacv1.Role
	if !ok {
		return fmt.Errorf("failed to decode RoleBinding spec")
	}

	_, err = client.RbacV1().RoleBindings(config.Kubernetes.JobsNamespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create RoleBinding: %v", err)
	}

	return nil
}

// DeleteRoleBinding deletes the RoleBinding created for a task.
func DeleteRoleBinding(ctx context.Context, taskID string, namespace string, client kubernetes.Interface, log *logger.Logger) error {
	rbs, err := client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=funnel,taskId=%s", taskID),
	})
	if err != nil {
		return fmt.Errorf("listing RoleBindings for task %s: %v", taskID, err)
	}
	for _, rb := range rbs.Items {
		log.Debug("deleting Worker RoleBinding", "name", rb.Name, "taskID", taskID)
		if err := client.RbacV1().RoleBindings(namespace).Delete(ctx, rb.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("deleting RoleBinding %s: %v", rb.Name, err)
		}
	}
	return nil
}
