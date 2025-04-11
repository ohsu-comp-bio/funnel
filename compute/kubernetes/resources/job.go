package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

// Create the Funnel Worker job from kubernetes-template.yaml
// Executor job is created in worker/kubernetes.go#Run
func CreateJob(task *tes.Task, namespace string, tpl string, client kubernetes.Interface, log *logger.Logger) error {
	// Parse Worker Template
	log.Debug("Creating job from template", "template", tpl)
	t, err := template.New(task.Id).Parse(tpl)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	res := task.GetResources()
	if res == nil {
		res = &tes.Resources{}
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"TaskId":    task.Id,
		"Namespace": namespace,
		"Cpus":      res.GetCpuCores(),
		"RamGb":     res.GetRamGb(),
		"DiskGb":    res.GetDiskGb(),
	})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	log.Debug("Job template: %s", buf.String())
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return err
	}

	job, ok := obj.(*v1.Job)
	if !ok {
		return fmt.Errorf("failed to decode job spec")
	}

	log.Debug("Creating job", "Job", job.Name, "Namespace", namespace)
	_, err = client.BatchV1().Jobs(namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// deleteJob removes deletes a kubernetes v1/batch job.
func DeleteJob(ctx context.Context, taskID string, client batchv1.JobInterface, log *logger.Logger) error {
	var gracePeriod int64 = 0
	var prop metav1.DeletionPropagation = metav1.DeletePropagationForeground

	err := client.Delete(ctx, taskID, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &prop,
	})

	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}
