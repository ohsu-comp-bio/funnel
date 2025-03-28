package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

func CreateConfigMap(taskId string, namespace string, conf config.Config, tpl string) error {
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
	})
	if err != nil {
		return fmt.Errorf("executing ConfigMap template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("decoding ConfigMap spec: %v", err)
	}

	// Create the ConfigMap that will contain the Funnel Worker Config (`funnel-worker.yaml`)
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("failed to decode ConfigMap spec")
	}

	cm.Data = make(map[string]string)

	configBytes, err := config.ToYaml(conf)
	if err != nil {
		return fmt.Errorf("marshaling config to ConfigMap: %v", err)
	}
	cm.Data["funnel-worker.yaml"] = string(configBytes)

	return nil
}

func DeleteConfigMap(taskId string, namespace string, client *kubernetes.Clientset) error {
	name := fmt.Sprintf("funnel-worker-%s", taskId)
	err := client.CoreV1().ConfigMaps(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})

	if err != nil {
		fmt.Errorf("deleting ConfigMap: %v", err)
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
