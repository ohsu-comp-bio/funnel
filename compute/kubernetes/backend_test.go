package kubernetes

import (
	"context"
	"fmt"
	"os"
	"testing"

	"path/filepath"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// getKubernetesClient creates a Kubernetes clientset for communication with the Kubernetes API
func getKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	var contextName string

	// Try to create an in-cluster client
	config, err = rest.InClusterConfig()
	if err != nil {
		// If in-cluster config fails, fallback to kubeconfig file for out-of-cluster
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		kubeConfig, err := clientcmd.LoadFromFile(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
		}

		// Get the current context name
		contextName = kubeConfig.CurrentContext

		// Build config from flags or kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
		}
	} else {
		contextName = "in-cluster"
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	// Log or print the current context being used
	fmt.Printf("Using Kubernetes context: %s\n", contextName)

	return clientset, nil
}

func TestCreateJobc(t *testing.T) {
	conf := config.DefaultConfig().Kubernetes
	content, err := os.ReadFile("../../config/kubernetes-template.yaml")
	if err != nil {
		t.Fatal(fmt.Errorf("reading template: %v", err))
	}
	conf.Template = string(content)
	log := logger.NewLogger("test", logger.DefaultConfig())

	// Create Kubernetes client
	clientset, err := getKubernetesClient()
	if err != nil {
		t.Fatal(fmt.Errorf("creating kubernetes client: %v", err))
	}

	b := &Backend{
		client:    clientset.BatchV1().Jobs(conf.Namespace),
		namespace: conf.Namespace,
		template:  conf.Template,
		event:     nil,
		database:  nil,
		log:       log,
	}

	task := &tes.Task{
		Id: "task1",
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	job, err := b.createJob(task)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", job)

	err = b.Submit(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
}
