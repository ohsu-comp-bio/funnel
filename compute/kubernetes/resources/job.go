package resources

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// Create the Funnel Worker job from kubernetes-template.yaml
// Executor job is created in worker/kubernetes.go#Run
func CreateJob(ctx context.Context, task *tes.Task, conf *config.Config, client kubernetes.Interface, log *logger.Logger) error {
	// Parse Worker Template

	log.Debug("Creating job from template", "template", conf.Kubernetes.WorkerTemplate)
	t, err := template.New(task.Id).Parse(conf.Kubernetes.WorkerTemplate)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	pods, err := client.CoreV1().Pods(conf.Kubernetes.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=funnel",
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	res := task.GetResources()
	if res == nil {
		res = &tes.Resources{}
	}

	templateData := map[string]interface{}{
		"TaskId":             task.Id,
		"Namespace":          conf.Kubernetes.Namespace,
		"JobsNamespace":      conf.Kubernetes.JobsNamespace,
		"Cpus":               res.GetCpuCores(),
		"RamGb":              res.GetRamGb(),
		"DiskGb":             res.GetDiskGb(),
		"Image":              pods.Items[0].Spec.Containers[0].Image,
		"NeedsPVC":           len(task.Inputs) > 0 || len(task.Outputs) > 0,
		"NodeSelector":       conf.Kubernetes.NodeSelector,
		"Tolerations":        conf.Kubernetes.Tolerations,
		"ServiceAccountName": fmt.Sprintf("funnel-worker-sa-%s-%s", conf.Kubernetes.JobsNamespace, task.Id),
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

	log.Debug("Job template", "template", buf.String())
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return err
	}

	job, ok := obj.(*v1.Job)
	if !ok {
		return fmt.Errorf("failed to decode job spec")
	}

	log.Debug("Creating job", "Job", job.Name, "JobsNamespace", conf.Kubernetes.JobsNamespace)
	_, err = client.BatchV1().Jobs(conf.Kubernetes.JobsNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

// DeleteJob removes the worker job for a task and all associated executor jobs.
func DeleteJob(ctx context.Context, conf *config.Config, taskID string, client kubernetes.Interface, log *logger.Logger) error {
	jobsInterface := client.BatchV1().Jobs(conf.Kubernetes.JobsNamespace)

	var gracePeriod int64 = 0
	var prop metav1.DeletionPropagation = metav1.DeletePropagationForeground

	// Delete the worker job (named exactly as taskID)
	err := jobsInterface.Delete(ctx, taskID, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &prop,
	})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("%v", err)
	}

	// Delete executor jobs (named as {taskID}-{index})
	return DeleteExecutorJobs(ctx, conf, taskID, client, log)
}

// DeleteExecutorJobs deletes all executor jobs associated with a task.
// Executor jobs are named {taskID}-{index} and are created by the worker process.
func DeleteExecutorJobs(ctx context.Context, conf *config.Config, taskID string, client kubernetes.Interface, log *logger.Logger) error {
	jobsInterface := client.BatchV1().Jobs(conf.Kubernetes.JobsNamespace)

	jobs, err := jobsInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing jobs to find executor jobs for task %s: %v", taskID, err)
	}

	prefix := taskID + "-"
	var errs error
	var gracePeriod int64 = 0
	var prop metav1.DeletionPropagation = metav1.DeletePropagationForeground

	for _, job := range jobs.Items {
		if len(job.Name) > len(prefix) && job.Name[:len(prefix)] == prefix {
			log.Debug("deleting executor job", "jobName", job.Name, "taskID", taskID)
			delErr := jobsInterface.Delete(ctx, job.Name, metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
				PropagationPolicy:  &prop,
			})
			if delErr != nil {
				errs = fmt.Errorf("deleting executor job %s: %v", job.Name, delErr)
			}
		}
	}

	return errs
}
