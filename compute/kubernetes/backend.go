// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/hashicorp/go-multierror"
	"github.com/ohsu-comp-bio/funnel/compute/kubernetes/resources"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/plugins/proto"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util/k8sutil"
)

// Backend represents the K8s backend.
type Backend struct {
	client            kubernetes.Interface
	event             events.Writer
	database          tes.ReadOnlyServer
	log               *logger.Logger
	backendParameters map[string]string
	conf              *config.Config // Funnel configuration
	events.Computer
}

// NewBackend returns a new K8s Backend instance.
func NewBackend(ctx context.Context, conf *config.Config, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	if conf.Kubernetes.WorkerTemplate == "" {
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
		client:   clientset,
		event:    writer,
		database: reader,
		log:      log,
		conf:     conf,
	}

	if !conf.Kubernetes.DisableReconciler {
		rate := conf.Kubernetes.ReconcileRate.AsDuration()
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

	var taskConfig *config.Config = b.conf
	if b.conf.Plugins != nil {
		resp, ok := ctx.Value("pluginResponse").(*proto.JobResponse)
		if !ok {
			return fmt.Errorf("Failed to unmarshal plugin response %v", ctx.Value("pluginResponse"))
		}
		taskConfig = resp.Config
	}

	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ctx, ev.GetTask(), taskConfig)
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
func (b *Backend) Submit(ctx context.Context, task *tes.Task, config *config.Config) error {
	err := b.createResources(task, config)
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
func (b *Backend) createResources(task *tes.Task, config *config.Config) error {
	// If the task has inputs or outputs that must be taken care of create a PVC
	if len(task.Inputs) > 0 || len(task.Outputs) > 0 {
		b.log.Debug("creating Worker PV", "taskID", task.Id)

		// Check to make sure required configs are present
		if config.GenericS3 == nil || len(config.GenericS3) == 0 ||
			config.GenericS3[0].Bucket == "" || config.GenericS3[0].Region == "" {
			return fmt.Errorf("Bucket or Region not found in GenericS3 config when attempting to create resources for task: %#v", task)
		}

		// Create PV
		err := resources.CreatePV(task.Id,
			config,
			b.client, b.log)
		if err != nil {
			return fmt.Errorf("creating Worker PV: %v", err)
		}

		// Create PVC
		b.log.Debug("creating Worker PVC", "taskID", task.Id)
		err = resources.CreatePVC(task.Id, config, b.client, b.log)
		if err != nil {
			return fmt.Errorf("creating Worker PVC: %v", err)
		}
	}

	// Create ConfigMap
	b.log.Debug("creating Worker ConfigMap", "taskID", task.Id)
	err := resources.CreateConfigMap(task.Id,
		config, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker ConfigMap: %v", err)
	}

	// Create Worker Job
	b.log.Debug("creating Worker Job", "taskID", task.Id)
	err = resources.CreateJob(task, config, b.client, b.log)
	if err != nil {
		return fmt.Errorf("creating Worker Job: %v", err)
	}

	return nil
}

// cleanResources deletes the resources created for a task.
func (b *Backend) cleanResources(ctx context.Context, taskId string) error {
	var errs error

	// Delete PV
	err := resources.DeletePV(ctx, taskId, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker PV: %v", err)
	}

	// Delete PVC
	err = resources.DeletePVC(ctx, taskId, b.conf.Kubernetes.JobsNamespace, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker PVC: %v", err)
	}

	// Delete ConfigMap
	err = resources.DeleteConfigMap(ctx, taskId, b.conf.Kubernetes.JobsNamespace, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Worker ConfigMap: %v", err)
	}

	//Delete Job
	b.log.Debug("deleting Job", "taskID", taskId)
	err = resources.DeleteJob(ctx, b.conf, taskId, b.client, b.log)
	if err != nil {
		errs = multierror.Append(errs, err)
		b.log.Error("deleting Job: %v", err)
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
	// Clears all resources that still exist from jobs that have run before it
	if !disableCleanup {
		jobs, err := b.client.BatchV1().Jobs(b.conf.Kubernetes.JobsNamespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			b.log.Error("backlog cleanup: listing jobs", err)
		} else {
			for _, j := range jobs.Items {
				s := j.Status
				if s.Succeeded > 0 || s.Failed > 0 {
					b.log.Debug("backlog cleanup: deleting job", "taskID", j.Name)
					if err := b.cleanResources(ctx, j.Name); err != nil {
						b.log.Error("backlog cleanup: failed to clean resources", "taskID", j.Name, "error", err)
					}
				}
			}
		}
	}

	ticker := time.NewTicker(rate)
	failedJobEvents := make(map[string]int)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			const maxErrEventWrites = 2

			jobs, err := b.client.BatchV1().Jobs(b.conf.Kubernetes.JobsNamespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				b.log.Error("reconcile: listing jobs", err)
				continue
			}
			for _, j := range jobs.Items {
				s := j.Status
				jobName := j.Name
				switch {
				case s.Active > 0:
					continue
				case s.Succeeded > 0:
					if disableCleanup {
						continue
					}
					b.log.Debug("reconcile: cleaning up successful job", "taskID", jobName)

					// Delete resources
					if err := b.cleanResources(ctx, jobName); err != nil {
						b.log.Error("failed to clean resources", "taskID", jobName, "error", err)
						continue
					}
					delete(failedJobEvents, jobName)

				case s.Failed > 0:
					if count, exists := failedJobEvents[jobName]; exists && count >= maxErrEventWrites {
						continue
					}

					b.log.Debug("reconcile: gathering failed k8s job conditions", "taskID", jobName)
					conds, err := json.Marshal(s.Conditions)
					if err != nil {
						b.log.Error("reconcile: marshal failed job conditions", "taskID", jobName, "error", err)
					}
					b.event.WriteEvent(ctx, events.NewState(jobName, tes.SystemError))
					b.event.WriteEvent(
						ctx,
						events.NewSystemLog(
							jobName, 0, 0, "error",
							"Kubernetes job in FAILED state",
							map[string]string{"error": string(conds)},
						),
					)

					failedJobEvents[jobName]++
					if disableCleanup {
						continue
					}

					b.log.Debug("reconcile: cleaning up failed job", "taskID", jobName)
					if err := b.cleanResources(ctx, jobName); err != nil {
						b.log.Error("failed to clean resources", "taskID", jobName, "error", err)
						continue
					}
					delete(failedJobEvents, jobName)
				}
			}
		}
	}
}
