package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/ohsu-comp-bio/funnel/tes"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// KubernetesCommand is responsible for configuring and running a task in a Kubernetes cluster.
type KubernetesCommand struct {
	TaskId          string
	JobId           int
	StdinFile       string
	TaskTemplate    string
	Namespace       string
	Resources       *tes.Resources
	Command
}

// Creates a new Kuberntes Job which will run the task.
func (kcmd KubernetesCommand) Run(ctx context.Context) error {
	var taskId = kcmd.TaskId 
	tpl, err := template.New(taskId).Parse(kcmd.TaskTemplate)

	if err != nil {
		return err
	}

	var command = kcmd.ShellCommand
	if kcmd.StdinFile != "" {
		command = append(command, "<", kcmd.StdinFile)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]interface{}{
		"TaskId":    taskId,
		"JobId":     kcmd.JobId,
		"Namespace": kcmd.Namespace,
		"Image":     kcmd.Image,
		"Command":   command,
		"Workdir":   kcmd.Workdir,
		"Volumes":   kcmd.Volumes,
		"Cpus":      kcmd.Resources.CpuCores,
		"RamGb":     kcmd.Resources.RamGb,
		"DiskGb":    kcmd.Resources.DiskGb,
	})

	if err != nil {
		return err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return err
	}

	job, ok := obj.(*v1.Job)
	if !ok {
		return err
	}

	clientset, err := getKubernetesClientset()
	if err != nil {
		return err
	}

	var client = clientset.BatchV1().Jobs(kcmd.Namespace)

	_ , err = client.Create(ctx, job, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("creating job: %v", err)
	}

	// Wait until the job finishes
	watcher, err := client.Watch(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s-%d", taskId, kcmd.JobId)})
	defer watcher.Stop()
	waitForJobFinnish(ctx, watcher)

	pods, err := clientset.CoreV1().Pods(kcmd.Namespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s-%d", taskId, kcmd.JobId)})
	if err != nil {
		return err
	}

	for _, v := range pods.Items {
		req := clientset.CoreV1().Pods(kcmd.Namespace).GetLogs(v.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(ctx) 

		if err != nil {
			return err
		}

		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return err 
		}

		var bytes = buf.Bytes()
		kcmd.Stdout.Write(bytes)
    }

	return nil
}

// Deletes the job running the task.
func (kcmd KubernetesCommand) Stop() error {
	clientset, err := getKubernetesClientset()
	if err != nil {
		return err
	}

	jobName := fmt.Sprintf("%s-%d", kcmd.TaskId, kcmd.JobId)

	backgroundDeletion := metav1.DeletePropagationBackground
	err = clientset.BatchV1().Jobs(kcmd.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{
		PropagationPolicy: &backgroundDeletion,
	})

	if err != nil {
		return fmt.Errorf("deleting job: %v", err)
	}

	return nil
}

// Waits until the job finishes
func waitForJobFinnish(ctx context.Context, watcher watch.Interface) {
	for {
		select {
		case event := <-watcher.ResultChan():
			job := event.Object.(*v1.Job)

			if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
				return
			} else if event.Type == watch.Deleted {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func getKubernetesClientset() (*kubernetes.Clientset, error) {
	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	return clientset, err
}