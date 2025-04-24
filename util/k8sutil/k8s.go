package k8sutil

import (
	"fmt"
	"os"

	"github.com/ohsu-comp-bio/funnel/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewK8sClient returns a new Kubernetes client.
func NewK8sClient(conf *config.Config) (*kubernetes.Clientset, error) {
	var kubeconfig *rest.Config
	var err error

	// Case 1: Use provided kubeconfig	 file
	if conf.Kubernetes.ConfigFile != "" {
		// use the current context in kubeconfig
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", conf.Kubernetes.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("building kubeconfig: %v", err)
		}
	} else if configPath := os.Getenv("KUBECONFIG"); configPath != "" {
		// Case 2: Fall back to KUBECONFIG env var or default kubeconfig
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, fmt.Errorf("building kubeconfig from env: %v", err)
		}
	} else {
		// Case 3: Fall back to in-cluster config
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

	return clientset, nil
}
