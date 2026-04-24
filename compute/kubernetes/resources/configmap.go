package resources

import (
	"context"
	"fmt"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateConfigMap(ctx context.Context, taskId string, conf *config.Config, client kubernetes.Interface, log *logger.Logger) error {
	configBytes, err := config.ToYaml(conf)
	if err != nil {
		return fmt.Errorf("marshaling config to ConfigMap: %v", err)
	}

	// Create the ConfigMap that will contain the Funnel Worker Config (`funnel-worker.yaml`)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("funnel-worker-config-%s", taskId),
			Namespace: conf.Kubernetes.JobsNamespace,
			Labels: map[string]string{
				"app":    "funnel",
				"taskId": taskId,
			},
		},
		Data: map[string]string{
			"funnel-worker.yaml": string(configBytes),
		},
	}

	_, err = client.CoreV1().ConfigMaps(conf.Kubernetes.JobsNamespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func DeleteConfigMap(ctx context.Context, taskId string, namespace string, client kubernetes.Interface, log *logger.Logger) error {
	name := fmt.Sprintf("funnel-worker-config-%s", taskId)
	_, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting ConfigMap %s: %v", name, err)
	}
	log.Debug("deleting Worker configMap", "taskID", taskId)
	if err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
