package resources

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

func CreateConfigMap(ctx context.Context, taskId string, conf *config.Config, client kubernetes.Interface, log *logger.Logger) error {
	configBytes, err := config.ToYaml(conf)
	if err != nil {
		return fmt.Errorf("marshaling config to ConfigMap: %v", err)
	}

	indentFn := func(spaces int, s string) string {
		pad := strings.Repeat(" ", spaces)
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = pad + line
			}
		}
		return strings.Join(lines, "\n")
	}

	t, err := template.New(taskId).Funcs(template.FuncMap{"indent": indentFn}).Parse(conf.Kubernetes.ConfigMapTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"TaskId":    taskId,
		"Namespace": conf.Kubernetes.JobsNamespace,
		"Data":      string(configBytes),
	})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	fmt.Println("DEBUG: ConfigMap YAML:\n", buf.String()) // Debugging line to print the generated YAML
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("decoding ConfigMap spec: %v", err)
	}

	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("failed to decode ConfigMap spec")
	}

	_, err = client.CoreV1().ConfigMaps(conf.Kubernetes.JobsNamespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func DeleteConfigMap(ctx context.Context, taskId string, namespace string, client kubernetes.Interface, log *logger.Logger) error {
	name := fmt.Sprintf("funnel-worker-config-%s", taskId)
	cfg, _ := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if cfg != nil {
		log.Debug("deleting Worker configMap", "taskID", taskId)
		err := client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}
