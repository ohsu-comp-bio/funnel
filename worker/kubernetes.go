package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

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
	TaskId         string
	JobId          int
	StdinFile      string
	TaskTemplate   string
	Namespace      string // Funnel Server Namespace
	JobsNamespace  string // Funnel Worker + Executor Namespace (default: Namespace)
	NodeSelector   map[string]string
	Tolerations    []map[string]interface{}
	Resources      *tes.Resources
	ServiceAccount string
	NeedsPVC       bool
	Clientset      kubernetes.Interface
	Command
}

// Utility function to correctly handle tasks with o/quotes in commands
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "\"" + strings.ReplaceAll(s, "'", `\\'`) + "\""
}

// Create the Executor K8s job from kubernetes-executor-template.yaml
// Funnel Worker job is created in compute/kubernetes/backend.go#CreateResources
func (kcmd KubernetesCommand) Run(ctx context.Context) error {
	var taskId = kcmd.TaskId
	tpl, err := template.New(taskId).Parse(kcmd.TaskTemplate)

	if err != nil {
		return err
	}

	var cmd = kcmd.ShellCommand
	if len(cmd) == 0 {
		return fmt.Errorf("Funnel Worker: No command specified for Executor.")
	}

	if kcmd.StdinFile != "" {
		cmd = append(cmd, "<", kcmd.StdinFile)
	}

	for i, v := range cmd {
		if strings.Contains(v, " ") {
			cmd[i] = shellQuote(v)
		}
	}

	templateData := map[string]interface{}{
		"TaskId":             taskId,
		"JobId":              kcmd.JobId,
		"Namespace":          kcmd.Namespace,
		"JobsNamespace":      kcmd.JobsNamespace,
		"Command":            cmd,
		"Workdir":            kcmd.Workdir,
		"Volumes":            kcmd.Volumes,
		"Cpus":               kcmd.Resources.CpuCores,
		"RamGb":              kcmd.Resources.RamGb,
		"DiskGb":             kcmd.Resources.DiskGb,
		"Image":              kcmd.Image,
		"NeedsPVC":           kcmd.NeedsPVC,
		"ServiceAccountName": kcmd.ServiceAccount,
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, templateData)

	if err != nil {
		return fmt.Errorf("Funnel Worker: failed to execute job template: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(buf.Bytes(), nil, nil)
	if err != nil {
		return fmt.Errorf("Funnel Worker: failed to decode job template: %v", err)
	}

	job, ok := obj.(*v1.Job)
	if !ok {
		return fmt.Errorf("Funnel Worker: decoded object is not a Job")
	}

	clientset := kcmd.Clientset
	if clientset == nil {
		var err error

		clientset, err = getKubernetesClientset()
		if err != nil {
			return fmt.Errorf("Funnel Worker: failed to get Kubernetes clientset: %v", err)
		}
	}

	var client = clientset.BatchV1().Jobs(kcmd.JobsNamespace)

	ret, err := client.Create(ctx, job, metav1.CreateOptions{})
	fmt.Println("DEBUG: Create returned:", ret)
	fmt.Println("DEBUG: Create err:", err)

	// Start log streaming and job completion monitoring in parallel
	errChan := make(chan error, 2)

	// Goroutine to handle log streaming
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in log streaming: %v", r)
			}
		}()

		// Try to get logs from executor pods
		logsRetrieved := false
		maxRetries := 10
		retryInterval := 2 * time.Second

		for i := 0; i < maxRetries && !logsRetrieved; i++ {
			// List pods for this job
			pods, err := clientset.CoreV1().Pods(kcmd.JobsNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s-%d", taskId, kcmd.JobId),
			})

			if err != nil {
				errChan <- fmt.Errorf("failed to list pods for executor job: %v", err)
				return
			}

			// Try to stream logs from each pod
			for _, pod := range pods.Items {
				fmt.Printf("DEBUG: Attempting to stream logs from pod %s\n", pod.Name)
				// Stream logs from any pod state (Running, Succeeded, Failed)
				// Kubernetes API supports fetching logs from terminated pods if they haven't been cleaned up yet
				err := streamPodLogsFromAnyState(ctx, kcmd.JobsNamespace, pod.Name, kcmd.Stdout)
				if err != nil {
					fmt.Printf("DEBUG: Failed to stream logs from pod %s: %v\n", pod.Name, err)
					// If streaming fails, continue to next pod or retry
					continue
				}
				fmt.Printf("DEBUG: Successfully streamed logs from pod %s\n", pod.Name)
				logsRetrieved = true
				break
			}

			if !logsRetrieved {
				time.Sleep(retryInterval)
			}
		}

		if !logsRetrieved {
			errChan <- fmt.Errorf("failed to retrieve logs from executor pods after %d attempts", maxRetries)
		} else {
			errChan <- nil
		}
	}()

	// Goroutine to wait for job completion
	go func() {
		watcher, err := client.Watch(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s-%d", taskId, kcmd.JobId)})
		if err != nil {
			errChan <- fmt.Errorf("failed to create job watcher: %v", err)
			return
		}
		defer watcher.Stop()

		waitForJobFinish(ctx, watcher)
		errChan <- nil
	}()

	// Wait for both log streaming and job completion
	var logErr, jobErr error
	completed := 0
	for completed < 2 {
		select {
		case err := <-errChan:
			if completed == 0 {
				logErr = err
			} else {
				jobErr = err
			}
			completed++
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// If log streaming failed but job completed, continue with job completion check
	// Log failure shouldn't prevent us from getting exit codes
	if logErr != nil && jobErr == nil {
		fmt.Printf("Warning: %v\n", logErr)
	} else if logErr != nil {
		return logErr
	}

	if jobErr != nil {
		return jobErr
	}

	jobName := fmt.Sprintf("%s-%d", taskId, kcmd.JobId)

	j, err := client.Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to retrieve final status for executor job %s: %v", jobName, err)
	}

	if j.Status.Failed > 0 {
		fmt.Printf("DEBUG: Job %s failed with %d failures, inspecting container status\n", jobName, j.Status.Failed)
		// Re-fetch pods to get final container state for actual exit codes and error details
		finalPods, err := clientset.CoreV1().Pods(kcmd.JobsNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s-%d", taskId, kcmd.JobId),
		})

		fmt.Printf("DEBUG: Found %d pods for job %s\n", len(finalPods.Items), jobName)
		if err != nil {
			fmt.Printf("DEBUG: Error listing pods: %v\n", err)
		}

		if err == nil && len(finalPods.Items) > 0 {
			pod := finalPods.Items[0]
			fmt.Printf("DEBUG: Pod %s phase: %s, containers: %d\n", pod.Name, pod.Status.Phase, len(pod.Status.ContainerStatuses))
			if len(pod.Status.ContainerStatuses) > 0 {
				cStatus := pod.Status.ContainerStatuses[0]
				fmt.Printf("DEBUG: Container state - Terminated: %v\n", cStatus.State.Terminated != nil)
				if cStatus.State.Terminated != nil {
					exitCode := int(cStatus.State.Terminated.ExitCode)
					message := cStatus.State.Terminated.Message
					reason := cStatus.State.Terminated.Reason
					fmt.Printf("DEBUG: Container terminated - ExitCode: %d, Reason: %s, Message: %s\n", exitCode, reason, message)

					// Include stderr if available from termination message
					if message != "" {
						return fmt.Errorf("executor job %s failed with exit code %d (%s): %s",
							jobName, exitCode, reason, message)
					}
					return fmt.Errorf("executor job %s failed with exit code %d (%s)",
						jobName, exitCode, reason)
				}
			} else {
				fmt.Printf("DEBUG: No container statuses found for pod %s\n", pod.Name)
			}
		} else {
			fmt.Printf("DEBUG: No pods found for job %s, err: %v\n", jobName, err)
		}

		// Fallback to original message if container inspection fails
		return fmt.Errorf("executor job %s failed with %d failures", jobName, j.Status.Failed)
	}

	return nil
}

func waitForPodRunning(ctx context.Context, namespace string, podName string, timeout time.Duration) (*corev1.Pod, error) {
	clientset, err := getKubernetesClientset()
	if err != nil {
		return nil, fmt.Errorf("failed getting kubernetes clientset: %v", err)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			return nil, fmt.Errorf("timed out waiting for pod %s to be in running state", podName)
		case <-ticker.C:
			pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("getting pod %s: %v", podName, err)
			}

			return pod, nil
		}
	}
}

func streamPodLogs(ctx context.Context, namespace string, podName string, stdout io.Writer) error {
	clientset, err := getKubernetesClientset()
	if err != nil {
		return fmt.Errorf("getting kubernetes clientset: %v", err)
	}

	// Stream stdout logs
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("streaming stdout logs: %v", err)
	}
	defer podLogs.Close()

	_, err = io.Copy(stdout, podLogs)
	return err
}

// streamPodLogsFromAnyState streams logs from a pod regardless of its state
// This works for Running, Succeeded, and Failed pods (as long as they haven't been deleted)
func streamPodLogsFromAnyState(ctx context.Context, namespace string, podName string, stdout io.Writer) error {
	clientset, err := getKubernetesClientset()
	if err != nil {
		return fmt.Errorf("getting kubernetes clientset: %v", err)
	}

	// Get logs from any pod state - Kubernetes API supports fetching logs from terminated pods
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		// Don't set Follow=true, we want to fetch all available logs, not stream live logs
	})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("streaming logs from pod %s: %v", podName, err)
	}
	defer podLogs.Close()

	_, err = io.Copy(stdout, podLogs)
	return err
}

// Deletes the job running the task.
func (kcmd KubernetesCommand) Stop() error {
	clientset, err := getKubernetesClientset()
	if err != nil {
		return err
	}

	jobName := fmt.Sprintf("%s-%d", kcmd.TaskId, kcmd.JobId)

	backgroundDeletion := metav1.DeletePropagationBackground
	err = clientset.BatchV1().Jobs(kcmd.JobsNamespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{
		PropagationPolicy: &backgroundDeletion,
	})

	if err != nil {
		return fmt.Errorf("deleting job: %v", err)
	}

	return nil
}

func (kcmd KubernetesCommand) GetStdout() io.Writer {
	return kcmd.Stdout
}

func (kcmd KubernetesCommand) GetStderr() io.Writer {
	return kcmd.Stderr
}

// Waits until the job finishes
func waitForJobFinish(ctx context.Context, watcher watch.Interface) {
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
