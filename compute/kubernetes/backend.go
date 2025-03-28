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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hashicorp/go-multierror"
	"github.com/ohsu-comp-bio/funnel/compute/kubernetes/resources"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// Backend represents the K8s backend.
type Backend struct {
	bucket            string
	region            string
	client            *kubernetes.Clientset
	namespace         string
	template          string
	pvTemplate        string
	pvcTemplate       string
	configMapTemplate string
	event             events.Writer
	database          tes.ReadOnlyServer
	log               *logger.Logger
	backendParameters map[string]string
	kubeconfig        *rest.Config  // Kubernetes client config
	conf              config.Config // Funnel configuration
	events.Computer
}

// NewBackend returns a new local Backend instance.
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
	if conf.Kubernetes.Namespace == "" {
		return nil, fmt.Errorf("invalid configuration; must provide a kubernetes namespace")
	}

	var kubeconfig *rest.Config
	var err error

	if conf.Kubernetes.ConfigFile != "" {
		// use the current context in kubeconfig
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", conf.Kubernetes.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("building kubeconfig: %v", err)
		}
	} else {
		// creates the in-cluster config
		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("building in-cluster kubeconfig: %v", err)
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		bucket:            conf.Kubernetes.Bucket,
		region:            conf.Kubernetes.Region,
		client:            clientset,
		namespace:         conf.Kubernetes.Namespace,
		template:          conf.Kubernetes.Template,
		pvTemplate:        conf.Kubernetes.PVTemplate,
		pvcTemplate:       conf.Kubernetes.PVCTemplate,
		configMapTemplate: conf.Kubernetes.ConfigMapTemplate,
		event:             writer,
		database:          reader,
		log:               log,
		kubeconfig:        kubeconfig, // Kubernetes client config
		conf:              conf,       // Funnel configuration
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
	if !b.conf.Plugins.Disabled {
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
		return fmt.Errorf("creating resources: %v", err)
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

func (b *Backend) createResources(task *tes.Task) error {
	// TODO: Update this so that a PVC/PV is only created if the task has inputs or outputs
	// If the task has either inputs or outputs, then create a PVC
	// shared between the Funnel Worker and the Executor
	// e.g. `if len(task.Inputs) > 0 || len(task.Outputs) > 0 {}`

	// Create PV
	err := resources.CreatePV(task.Id, b.namespace, b.bucket, b.region, b.pvTemplate)
	if err != nil {
		return fmt.Errorf("creating job: %v", err)
	}

	// Create PVC
	err = resources.CreatePVC(task.Id, b.namespace, b.bucket, b.region, b.pvcTemplate)
	if err != nil {
		return fmt.Errorf("creating PVC: %v", err)
	}

	// Create ConfigMap
	err = resources.CreateConfigMap(task.Id, b.namespace, b.conf, b.configMapTemplate)
	if err != nil {
		return fmt.Errorf("creating ConfigMap: %v", err)
	}

	// Create Job
	err = resources.CreateJob(task, b.namespace, b.template)
	if err != nil {
		return fmt.Errorf("creating job Job: %v", err)
	}

	return nil
}

func (b *Backend) cleanResources(ctx context.Context, taskId string) error {
	var errs error

	// Create PV
	err := resources.DeletePV(ctx, taskId, b.client)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting PV: %v", err)
	}

	// Create PVC
	err = resources.DeletePVC(ctx, taskId, b.namespace, b.client)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting PVC: %v", err)
	}

	// Create ConfigMap
	err = resources.DeleteConfigMap(ctx, taskId, b.namespace, b.client)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting ConfigMap: %v", err)
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
			jobs, err := b.client.BatchV1().Jobs(b.namespace).List(ctx, metav1.ListOptions{})
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
