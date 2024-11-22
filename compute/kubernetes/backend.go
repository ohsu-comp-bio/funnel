// Package kubernetes contains code for accessing compute resources via the Kubernetes v1 Batch API.
package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/template"
	"time"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// NewBackend returns a new local Backend instance.
func NewBackend(ctx context.Context, conf config.Kubernetes, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	if conf.TemplateFile != "" {
		content, err := os.ReadFile(conf.TemplateFile)
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
		config:    kubeconfig,
	}

	if !conf.DisableReconciler {
		rate := time.Duration(conf.ReconcileRate)
		go b.reconcile(ctx, rate, conf.DisableJobCleanup)
	}

	return b, nil
}

// Backend represents the local backend.
type Backend struct {
	client            batchv1.JobInterface
	namespace         string
	template          string
	event             events.Writer
	database          tes.ReadOnlyServer
	log               *logger.Logger
	backendParameters map[string]string
	config            *rest.Config
	events.Computer
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

// Create the Funnel Worker job
// Executor job is created in worker/kubernetes.go#Run
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

func (b *Backend) createPVC(ctx context.Context, taskID string, resources *tes.Resources) error {
	clientset, err := kubernetes.NewForConfig(b.config)
	if err != nil {
		return fmt.Errorf("getting kubernetes client: %v", err)
	}

	// Define storage size (ignored by S3 CSI driver but required by the API)
	storageSize := resource.NewQuantity(1024*1024*1024, resource.BinarySI) // Default 1Gi
	if resources != nil && resources.DiskGb > 0 {
		storageSize = resource.NewQuantity(int64(resources.DiskGb*1024*1024*1024), resource.BinarySI)
	}

	pvName := fmt.Sprintf("funnel-pv-%s", taskID)
	pvcName := fmt.Sprintf("funnel-pvc-%s", taskID)
	bucketName := "funnel-testing"

	// Step 1: Create the PersistentVolume
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
			Labels: map[string]string{
				"app":    "funnel",
				"taskId": taskID,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: *storageSize,
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany, // S3 CSI supports RWX
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              "",
			ClaimRef: &corev1.ObjectReference{
				Namespace: b.namespace,
				Name:      pvcName,
			},
			MountOptions: []string{
				"allow-delete",
				"region=us-west-2",
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver:       "s3.csi.aws.com",
					VolumeHandle: fmt.Sprintf("s3-csi-%s", taskID),
					VolumeAttributes: map[string]string{
						"bucketName": bucketName,
					},
				},
			},
		},
	}

	if _, err := clientset.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("creating PersistentVolume: %v", err)
	}

	// Step 2: Create the PersistentVolumeClaim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
			Labels: map[string]string{
				"app":    "funnel",
				"taskId": taskID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *storageSize,
				},
			},
			StorageClassName: func() *string { s := ""; return &s }(),
			VolumeName:       pvName,
		},
	}

	if _, err := clientset.CoreV1().PersistentVolumeClaims(b.namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("creating PersistentVolumeClaim: %v", err)
	}

	return nil
}

func (b *Backend) addOwnerReference(ctx context.Context, taskID string, job *v1.Job) error {
	// Fetch the Job that will own the PVC and PV
	jobName := fmt.Sprintf("funnel-job-%s", taskID)
	_, err := b.client.Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("fetching Job for owner reference: %v", err)
	}

	// Fetch the PVC
	pvcName := fmt.Sprintf("funnel-pvc-%s", taskID)
	clientset, err := kubernetes.NewForConfig(b.config)
	if err != nil {
		return fmt.Errorf("getting kubernetes client: %v", err)
	}
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(b.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("fetching PVC: %v", err)
	}

	// Add OwnerReference for the Job
	ownerRef := metav1.OwnerReference{
		APIVersion:         "batch/v1",
		Kind:               "Job",
		Name:               job.Name,
		UID:                job.UID,
		BlockOwnerDeletion: ptr.To(true),
		Controller:         ptr.To(true),
	}

	// Append the OwnerReference to the PVC
	pvc.OwnerReferences = append(pvc.OwnerReferences, ownerRef)

	// Update the PVC
	_, err = clientset.CoreV1().PersistentVolumeClaims(b.namespace).Update(ctx, pvc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating PVC: %v", err)
	}

	return nil
}

// Add this helper function for PVC cleanup
func (b *Backend) deletePVC(ctx context.Context, taskID string) error {
	clientset, err := kubernetes.NewForConfig(b.config)
	if err != nil {
		return fmt.Errorf("getting kubernetes client: %v", err)
	}

	pvcName := fmt.Sprintf("funnel-pvc-%s", taskID)
	err = clientset.CoreV1().PersistentVolumeClaims(b.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		// If the PVC is already gone, ignore the error
		if k8errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("deleting shared PVC: %v", err)
	}

	return nil
}

// Submit creates both the PVC and the worker job with better error handling
func (b *Backend) Submit(ctx context.Context, task *tes.Task) error {
	// Create a new background context instead of inheriting from the potentially canceled one
	submitCtx := context.Background()

	// TODO: Update this so that a PVC is only created if the task has inputs or outputs
	// If the task has either inputs or outputs, then create a PVC
	// shared between the Funnel Worker and the Executor
	// e.g. `if len(task.Inputs) > 0 || len(task.Outputs) > 0 {}`
	err := b.createPVC(submitCtx, task.Id, task.GetResources())
	if err != nil {
		return fmt.Errorf("creating shared storage: %v", err)
	}

	// Create the worker job
	job, err := b.createJob(task)
	if err != nil {
		return fmt.Errorf("creating job spec: %v", err)
	}

	_, err = b.client.Create(submitCtx, job, metav1.CreateOptions{
		FieldManager: task.Id,
	})
	if err != nil {
		return fmt.Errorf("creating job in backend: %v", err)
	}

	err = b.addOwnerReference(submitCtx, task.Id, job)
	if err != nil {
		return fmt.Errorf("updating PVC with OwnerReference: %v", err)
	}

	return nil
}

// deleteJob removes deletes a kubernetes v1/batch job.
func (b *Backend) deleteJob(ctx context.Context, taskID string) error {
	var gracePeriod int64 = 0
	var prop metav1.DeletionPropagation = metav1.DeletePropagationForeground
	err := b.client.Delete(ctx, taskID, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &prop,
	})
	if err != nil {
		return fmt.Errorf("deleting job: %v", err)
	}

	// Delete Worker PVC
	if err := b.deletePVC(ctx, taskID); err != nil {
		b.log.Error("failed to delete PVC", "taskID", taskID, "error", err)
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

	return b.deleteJob(ctx, taskID)
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
			jobs, err := b.client.List(ctx, metav1.ListOptions{})
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
					if err := b.deletePVC(ctx, j.Name); err != nil {
						b.log.Error("failed to delete PVC", "taskID", j.Name, "error", err)
					}

					err := b.deleteJob(ctx, j.Name)
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

					// Delete Worker PVC
					if err := b.deletePVC(ctx, j.Name); err != nil {
						b.log.Error("reconcile: cleaning up PVC for failed job", "taskID", j.Name, "error", err)
					}

					err = b.deleteJob(ctx, j.Name)
					if err != nil {
						b.log.Error("reconcile: cleaning up failed job", "taskID", j.Name, "error", err)
						continue ReconcileLoop
					}
				}
			}
		}
	}
}
