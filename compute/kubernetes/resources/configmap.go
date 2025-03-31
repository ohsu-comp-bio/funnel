package resources

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateConfigMap(taskId string, namespace string, conf config.Config, client kubernetes.Interface) error {
	configBytes, err := config.ToYaml(conf)
	if err != nil {
		return fmt.Errorf("marshaling config to ConfigMap: %v", err)
	}

	// Create the ConfigMap that will contain the Funnel Worker Config (`funnel-worker.yaml`)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("funnel-worker-config-%s", taskId),
			Namespace: namespace,
		},
		Data: map[string]string{
			"funnel-worker.yaml": string(configBytes),
		},
	}

	_, err = client.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating ConfigMap: %v", err)
	}

	return nil
}

func DeleteConfigMap(ctx context.Context, taskId string, namespace string, client kubernetes.Interface) error {
	name := fmt.Sprintf("funnel-worker-config-%s", taskId)
	err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("deleting ConfigMap: %v", err)
	}

	return nil
}

func UpdateConfig(ctx context.Context, dst *config.Config) error {
	resp, ok := ctx.Value("pluginResponse").(*shared.Response)
	if !ok {
		return fmt.Errorf("Failed to unmarshal plugin response %v", ctx.Value("pluginResponse"))
	}

	// TODO: Review all security implications of merging configs (injections, shared state, etc.)
	if resp.Config != nil {
		err := MergeConfigs(dst, resp.Config)
		if err != nil {
			return fmt.Errorf("error merging configs from plugin %v", err)
		}
	}

	return nil
}

func MergeConfigs(dst *config.Config, src *config.Config) error {
	err := mergo.MergeWithOverwrite(dst, src)
	if err != nil {
		return fmt.Errorf("error merging configs: %v", err)
	}

	return nil
}
