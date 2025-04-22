// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/hashicorp/go-multierror"
	"github.com/ohsu-comp-bio/funnel/compute/kubernetes/resources"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util/k8sutil"
)

// Backend represents the K8s backend.
type Backend struct {
	bucket            string
	region            string
	client            kubernetes.Interface
	namespace         string
	jobsNamespace     string
	template          string
	pvTemplate        string
	pvcTemplate       string
	configMapTemplate string
	event             events.Writer
	database          tes.ReadOnlyServer
	log               *logger.Logger
	backendParameters map[string]string
	conf              config.Config // Funnel configuration
	events.Computer
}

// NewBackend returns a new K8s Backend instance.
func NewBackend(ctx context.Context, conf config.Config, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	if conf.Kubernetes.TemplateFile != "" {
		content, err := os.ReadFile(conf.Kubernetes.TemplateFile)
		if err != nil {
			return nil, fmt.Errorf("reading template: %v", err)
		}
		conf.Kubernetes.Template = string(content)
	}
	if conf.Kubernetes.Template == "" {
		return nil, fmt.Errorf("invalid configuration; must provide a kubernetes job template")
	}

	// Funnel Server Namespace
	if conf.Kubernetes.Namespace == "" {
		return nil, fmt.Errorf("invalid configuration; must provide a kubernetes namespace")
	}

	// Funnel Worker + Executor Namespace
	if conf.Kubernetes.JobsNamespace == "" {
		conf.Kubernetes.JobsNamespace = conf.Kubernetes.Namespace
	}

	clientset, err := k8sutil.NewK8sClient(conf)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %v", err)
	}

	b := &Backend{
		bucket:            conf.Kubernetes.Bucket,
		region:            conf.Kubernetes.Region,
		client:            clientset,
		namespace:         conf.Kubernetes.Namespace,
		jobsNamespace:     conf.Kubernetes.JobsNamespace,
		template:          conf.Kubernetes.Template,
		pvTemplate:        conf.Kubernetes.PVTemplate,
		pvcTemplate:       conf.Kubernetes.PVCTemplate,
		configMapTemplate: conf.Kubernetes.ConfigMapTemplate,
		event:             writer,
		database:          reader,
		log:               log,
		conf:              conf, // Funnel configuration
	}

	if !conf.Kubernetes.DisableReconciler {
		rate := time.Duration(conf.Kubernetes.ReconcileRate)
		go b.reconcile(ctx, rate, conf.Kubernetes.DisableJobCleanup)
	}

	return b, nil
}

func (b Backend) CheckBackendParameterSupport(task *tes.Task) error {
	if !task.Resources.GetBackendParametersStrict() {
		return nil
	}

	taskBackendParameters := task.Resources.GetBackendParameters()
	for k := range taskBackendParameters {
		_, ok := b.backendParameters[k]
		if !ok {
			return errors.New("backend parameters not supported")
		}
	}

	return nil
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	// TODO: Should this be moved to the switch statement so it's only run on TASK_CREATED?
	if b.conf.Plugins != nil {
		err := resources.UpdateConfig(ctx, &b.conf)
		if err != nil {
			return fmt.Errorf("error updating config from plugin response: %v", err)
		}
	}

	switch ev.Type {
	case events.Type_TASK_CREATED:

		return b.Submit(ctx, ev.GetTask())

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

// Submit creates both the PVC and the worker job with better error handling
func (b *Backend) Submit(ctx context.Context, task *tes.Task) error {
	err := b.createResources(task)
	if err != nil {
		return fmt.Errorf("creating Worker resources: %v", err)
	}

	return nil
}

// Cancel removes tasks that are pending kubernetes v1/batch jobs.
func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	task, err := b.database.GetTask(
		ctx, &tes.GetTaskRequest{Id: taskID, View: tes.View_MINIMAL.String()},
	)
	if err != nil {
		return err
	}

	// only cancel tasks in a QUEUED state
	if task.State != tes.State_QUEUED {
		return nil
	}

	return b.cleanResources(ctx, taskID)
}

// createResources creates the resources needed for a task.
func (b *Backend) createResources(task *tes.Task) error {
	// TODO: Update this so that a PVC/PV is only created if the task has inputs or outputs
	// If the task has either inputs or outputs, then create a PVC
	// shared between the Funnel Worker and the Executor
	// e.g. `if len(task.Inputs) > 0 || len(task.Outputs) > 0 {...}`

	// Create PV
	b.log.Debug("creating Worker PV", "taskID", task.Id)
	err := resources.CreatePV(task.Id, b.jobsNamespace, b.bucket, b.region, b.pvTemplate, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker PV: %v", err)
	}

	// Create PVC
	b.log.Debug("creating Worker PVC", "taskID", task.Id)
	err = resources.CreatePVC(task.Id, b.jobsNamespace, b.bucket, b.region, b.pvcTemplate, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker PVC: %v", err)
	}

	// Create ConfigMap
	b.log.Debug("creating Worker ConfigMap", "taskID", task.Id)
	err = resources.CreateConfigMap(task.Id, b.jobsNamespace, b.conf, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker ConfigMap: %v", err)
	}

	// Create Worker Job
	b.log.Debug("creating Worker Job", "taskID", task.Id)
	err = resources.CreateJob(task, b.namespace, b.jobsNamespace, b.template, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker Job: %v", err)
	}

	return nil
}

// cleanResources deletes the resources created for a task.
func (b *Backend) cleanResources(ctx context.Context, taskId string) error {
	var errs error

	// Delete PV
	b.log.Debug("deleting Worker PV", "taskID", taskId)
	err := resources.DeletePV(ctx, taskId, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker PV: %v", err)
	}

	// Delete PVC
	b.log.Debug("deleting Worker PVC", "taskID", taskId)
	err = resources.DeletePVC(ctx, taskId, b.jobsNamespace, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker PVC: %v", err)
	}

	// Delete ConfigMap
	b.log.Debug("deleting Worker ConfigMap", "taskID", taskId)
	err = resources.DeleteConfigMap(ctx, taskId, b.jobsNamespace, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker ConfigMap: %v", err)
	}

	return errs
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
			jobs, err := b.client.BatchV1().Jobs(b.jobsNamespace).List(ctx, metav1.ListOptions{})
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

					// Delete Worker PVC
					if err := b.cleanResources(ctx, j.Name); err != nil {
						b.log.Error("failed to clean resources", "taskID", j.Name, "error", err)
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

					if err := b.cleanResources(ctx, j.Name); err != nil {
						b.log.Error("failed to clean resources", "taskID", j.Name, "error", err)
						continue ReconcileLoop
					}
				}
			}
		}
	}
}
