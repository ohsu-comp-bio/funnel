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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
	"github.com/ohsu-comp-bio/funnel/tes"
)

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
		client:            clientset.BatchV1().Jobs(conf.Kubernetes.Namespace),
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

// Backend represents the K8s backend.
type Backend struct {
	bucket            string
	region            string
	client            batchv1.JobInterface
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
		err := b.UpdateConfig(ctx)
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

func (b *Backend) UpdateConfig(ctx context.Context) error {
	resp, ok := ctx.Value("pluginResponse").(*shared.Response)
	if !ok {
		return fmt.Errorf("Failed to unmarshal plugin response %v", ctx.Value("pluginResponse"))
	}

	// TODO: Review all security implications of merging configs (injections, shared state, etc.)
	if resp.Config != nil {
		err := b.MergeConfigs(*resp.Config)
		if err != nil {
			b.log.Error("error merging configs from plugin", "error", err)
		}
	}

	return nil
}

func (b *Backend) MergeConfigs(c config.Config) error {
	err := mergo.MergeWithOverwrite(&b.conf, c)
	if err != nil {
		return fmt.Errorf("error merging configs: %v", err)
	}

	return nil
}

func (b *Backend) Close() {
	//TODO: close database?
}

func (b *Backend) createConfigMap(task *tes.Task) (*corev1.ConfigMap, error) {
	// Load templates
	cmTpl, err := template.New(task.Id).Parse(b.conf.Kubernetes.ConfigMapTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	var buf bytes.Buffer
	err = cmTpl.Execute(&buf, map[string]interface{}{
		"TaskId":    task.Id,
		"Namespace": b.namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("executing ConfigMap template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding ConfigMap spec: %v", err)
	}

	// Create the ConfigMap that will contain the Funnel Worker Config (`funnel-worker.yaml`)
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("failed to decode ConfigMap spec")
	}

	cm.Data = make(map[string]string)

	// TODO: This is a hack to be able to support dynamic namespaces
	// How can we support this in a more configurable way?
	// e.g. `funnel-worker.yaml` should be able to reference whatever namespace the Funnel server is in
	b.conf.Server.HostName = fmt.Sprintf("funnel.%v.svc.cluster.local", b.namespace)

	configBytes, err := config.ToYaml(b.conf)
	if err != nil {
		return nil, fmt.Errorf("marshaling config to ConfigMap: %v", err)
	}
	cm.Data["funnel-worker.yaml"] = string(configBytes)

	return cm, nil
}

// Create the Funnel Worker job from kubernetes-template.yaml
// Executor job is created in worker/kubernetes.go#Run
func (b *Backend) createJob(task *tes.Task) (*v1.Job, error) {
	// Parse Worker Template
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
		return nil, fmt.Errorf("executing Worker template: %v", err)
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

// Create the Worker/Executor PVC from config/kubernetes-pvc.yaml
// TODO: Move this config file to Helm Charts so users can see/customize it
func (b *Backend) createPVC(task *tes.Task) (*corev1.PersistentVolumeClaim, error) {
	// Load templates
	pvcTpl, err := template.New(task.Id).Parse(b.pvcTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	var buf bytes.Buffer
	err = pvcTpl.Execute(&buf, map[string]interface{}{
		"TaskId":    task.Id,
		"Namespace": b.namespace,
		"Bucket":    b.bucket,
		"Region":    b.region,
	})
	if err != nil {
		return nil, fmt.Errorf("executing PVC template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding PVC spec: %v", err)
	}

	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return nil, fmt.Errorf("failed to decode PVC spec")
	}
	return pvc, nil
}

// Create the Worker/Executor PV from config/kubernetes-pv.yaml
// TODO: Move this config file to Helm Charts so users can see/customize it
func (b *Backend) createPV(task *tes.Task) (*corev1.PersistentVolume, error) {
	// Load templates
	pvTpl, err := template.New(task.Id).Parse(b.pvTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	// Template parameters
	var buf bytes.Buffer
	err = pvTpl.Execute(&buf, map[string]interface{}{
		"TaskId":    task.Id,
		"Namespace": b.namespace,
		"Bucket":    b.bucket,
		"Region":    b.region,
	})
	if err != nil {
		return nil, fmt.Errorf("executing PV template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding PV spec: %v", err)
	}

	pv, ok := obj.(*corev1.PersistentVolume)
	if !ok {
		return nil, fmt.Errorf("failed to decode PV spec")
	}
	return pv, nil
}

// Add this helper function for PVC cleanup
func (b *Backend) deletePVC(ctx context.Context, taskID string) error {
	clientset, err := kubernetes.NewForConfig(b.kubeconfig)
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
	clientset, err := kubernetes.NewForConfig(b.kubeconfig)
	if err != nil {
		return fmt.Errorf("getting kubernetes client: %v", err)
	}

	// Initialize ConfigMap
	cm, err := b.createConfigMap(task)
	if err != nil {
		return fmt.Errorf("initializing ConfigMap: %v", err)
	}

	// Create ConfigMap
	cm, err = clientset.CoreV1().ConfigMaps(b.namespace).Create(context.Background(), cm, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating ConfigMap: %v", err)
	}

	// TODO: Update this so that a PVC/PV is only created if the task has inputs or outputs
	// If the task has either inputs or outputs, then create a PVC
	// shared between the Funnel Worker and the Executor
	// e.g. `if len(task.Inputs) > 0 || len(task.Outputs) > 0 {}`

	// Initialize PV
	pv, err := b.createPV(task)
	if err != nil {
		return fmt.Errorf("initiaizing PV: %v", err)
	}

	// Create PV
	pv, err = clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating PV: %v", err)
	}

	// Initialize PVC
	pvc, err := b.createPVC(task)
	if err != nil {
		return fmt.Errorf("initializing PVC: %v", err)
	}

	// Create PVC
	pvc, err = clientset.CoreV1().PersistentVolumeClaims(b.namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating PVC: %v", err)
	}

	// Initialize Worker Job
	job, err := b.createJob(task)
	if err != nil {
		return fmt.Errorf("creating job spec: %v", err)
	}

	// Create Worker Job
	_, err = b.client.Create(context.Background(), job, metav1.CreateOptions{
		FieldManager: task.Id,
	})
	if err != nil {
		return fmt.Errorf("creating job in backend: %v", err)
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
