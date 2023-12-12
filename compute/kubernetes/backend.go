// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"
	"time"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, conf config.Kubernetes, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	if conf.TemplateFile != "" {
		content, err := ioutil.ReadFile(conf.TemplateFile)
		if err != nil {
			return nil, fmt.Errorf("reading template: %v", err)
		}
		conf.Template = string(content)
	}
	if conf.Template == "" {
		return nil, fmt.Errorf("invalid configuration; must provide a kubernetes job template")
	}
	if conf.Namespace == "" {
		return nil, fmt.Errorf("invalid configuration; must provide a kubernetes namespace")
	}

	var kubeconfig *rest.Config
	var err error

	if conf.ConfigFile != "" {
		// use the current context in kubeconfig
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", conf.ConfigFile)
		if err != nil {
			return nil, err
		}
	} else {
		// creates the in-cluster config
		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		client:    clientset.BatchV1().Jobs(conf.Namespace),
		namespace: conf.Namespace,
		template:  conf.Template,
		event:     writer,
		database:  reader,
		log:       log,
	}

	if !conf.DisableReconciler {
		rate := time.Duration(conf.ReconcileRate)
		go b.reconcile(ctx, rate, conf.DisableJobCleanup)
	}

	return b, nil
}

// Backend represents the local backend.
type Backend struct {
	client    batchv1.JobInterface
	namespace string
	template  string
	event     events.Writer
	database  tes.ReadOnlyServer
	log       *logger.Logger
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ctx, ev.Id)
		}
	}
	return nil
}

func (b *Backend) Close() {
	//TODO: close database?
}

// createJob uses the configured template to create a kubernetes batch job.
func (b *Backend) createJob(task *tes.Task) (*v1.Job, error) {
	submitTpl, err := template.New(task.Id).Parse(b.template)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	res := task.GetResources()
	if res == nil {
		res = &tes.Resources{}
	}

	var buf bytes.Buffer
	err = submitTpl.Execute(&buf, map[string]interface{}{
		"TaskId":    task.Id,
		"Namespace": b.namespace,
		"Cpus":      res.GetCpuCores(),
		"RamGb":     res.GetRamGb(),
		"DiskGb":    res.GetDiskGb(),
	})
	if err != nil {
		return nil, fmt.Errorf("executing template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding job spec: %v", err)
	}

	job, ok := obj.(*v1.Job)
	if !ok {
		return nil, fmt.Errorf("failed to decode job spec")
	}
	return job, nil
}

// Submit submits a task to the as a kubernetes v1/batch job.
func (b *Backend) Submit(task *tes.Task) error {
	job, err := b.createJob(task)
	if err != nil {
		return fmt.Errorf("creating job spec: %v", err)
	}
	_, err = b.client.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating job: %v", err)
	}
	return nil
}

// deleteJob removes deletes a kubernetes v1/batch job.
func (b *Backend) deleteJob(taskID string) error {
	var gracePeriod int64 = 0
	var prop metav1.DeletionPropagation = metav1.DeletePropagationForeground
	err := b.client.Delete(context.TODO(), taskID, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &prop,
	})
	if err != nil {
		return fmt.Errorf("deleting job: %v", err)
	}
	return nil
}

// Cancel removes tasks that are pending kubernetes v1/batch jobs.
func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	task, err := b.database.GetTask(
		ctx, &tes.GetTaskRequest{Id: taskID, View: tes.TaskView_MINIMAL},
	)
	if err != nil {
		return err
	}

	// only cancel tasks in a QUEUED state
	if task.State != tes.State_QUEUED {
		return nil
	}

	return b.deleteJob(taskID)
}

// Reconcile loops through tasks and checks the status from Funnel's database
// against the status reported by Kubernetes. This allows the backend to report
// system error's that prevented the worker process from running.
//
// Currently this handles a narrow set of cases:
//
// |---------------------|-----------------|--------------------|
// |    Funnel State     |  Backend State  |  Reconciled State  |
// |---------------------|-----------------|--------------------|
// |        QUEUED       |     FAILED      |    SYSTEM_ERROR    |
// |  INITIALIZING       |     FAILED      |    SYSTEM_ERROR    |
// |       RUNNING       |     FAILED      |    SYSTEM_ERROR    |
//
// In this context a "FAILED" state is being used as a generic term that captures
// one or more terminal states for the backend.
//
// This loop is also used to cleanup successful jobs.
func (b *Backend) reconcile(ctx context.Context, rate time.Duration, disableCleanup bool) {
	ticker := time.NewTicker(rate)
ReconcileLoop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobs, err := b.client.List(context.TODO() ,metav1.ListOptions{})
			if err != nil {
				b.log.Error("reconcile: listing jobs", err)
				continue ReconcileLoop
			}
			for _, j := range jobs.Items {
				s := j.Status
				switch {
				case s.Succeeded > 0:
					if disableCleanup {
						continue ReconcileLoop
					}
					b.log.Debug("reconcile: cleanuping up successful job", "taskID", j.Name)
					err := b.deleteJob(j.Name)
					if err != nil {
						b.log.Error("reconcile: cleaning up successful job", "taskID", j.Name, "error", err)
						continue ReconcileLoop
					}
				case s.Failed > 0:
					b.log.Debug("reconcile: cleaning up failed job", "taskID", j.Name)
					conds, err := json.Marshal(s.Conditions)
					if err != nil {
						b.log.Error("reconcile: marshal failed job conditions", "taskID", j.Name, "error", err)
					}
					b.event.WriteEvent(ctx, events.NewState(j.Name, tes.SystemError))
					b.event.WriteEvent(
						ctx,
						events.NewSystemLog(
							j.Name, 0, 0, "error",
							"Kubernetes job in FAILED state",
							map[string]string{"error": string(conds)},
						),
					)
					if disableCleanup {
						continue ReconcileLoop
					}
					err = b.deleteJob(j.Name)
					if err != nil {
						b.log.Error("reconcile: cleaning up failed job", "taskID", j.Name, "error", err)
						continue ReconcileLoop
					}
				}
			}
		}
	}
}
