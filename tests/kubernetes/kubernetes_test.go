package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
)

var fun *tests.Funnel
var clusterName string

// checkDependencies checks if the required dependencies (kind and helm) are installed.
// Kind: https://kind.sigs.k8s.io/
// Helm: https://helm.sh/
func checkDependencies() error {
	cmd := exec.Command("kind", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kind command not found: %v", err)
	}

	cmd = exec.Command("helm", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm command not found: %v", err)
	}

	return nil
}

// TestMain sets up the test environment for Kubernetes integration tests.
func TestMain(m *testing.M) {
	if err := checkDependencies(); err != nil {
		logger.Debug("Test dependencies not found:", err)
		os.Exit(0)
	}

	tests.ParseConfig()
	conf := tests.DefaultConfig()

	// Set the compute type to Kubernetes
	conf.Compute = "kubernetes"

	// Set the cluster name
	clusterName = "funnel-test-cluster-" + tests.RandomString(6)

	// Set the kubeconfig file path
	kubeconfig := filepath.Join(os.TempDir(), "funnel", fmt.Sprintf("%s-kubeconfig.yaml", clusterName))
	conf.Kubernetes.ConfigFile = kubeconfig

	// TODO: This pattern is used when running tests from the Makefile (e.g. `make test-slurm`)
	// We don't have any `test-k8s` target, but can add one if needed.
	// Currently, we'r running theese tests directly either in VS Code or `go test ./tests/kubernetes`
	if conf.Compute != "kubernetes" {
		logger.Debug("Skipping kubernetes e2e tests...")
		os.Exit(0)
	}

	err := StartK8sCluster(clusterName, kubeconfig)
	if err != nil {
		logger.Debug("failed to start Kind cluster:", err)
		os.Exit(1)
	}

	err = StartServerInK8s(clusterName, "../../deployments/kubernetes/helm/")
	if err != nil {
		logger.Debug("failed to start funnel server in k8s:", err)
		os.Exit(1)
	}

	exit := 0
	defer func() {
		DeleteK8sCluster(clusterName, kubeconfig)
		os.Exit(exit)
	}()

	exit = m.Run()
	return
}

// StartK8sCluster creates a kind cluster for integration tests.
func StartK8sCluster(clusterName string, kubeconfig string) error {
	cmd := exec.Command("kind", "create", "cluster",
		"--name", clusterName,
		"--kubeconfig", kubeconfig,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Kind cluster: %v", err)
	}

	// Wait a bit for the API server to settle
	logger.Debug("Kind cluster %q created with kubeconfig at: %s", clusterName, kubeconfig)

	return nil
}

// DeleteK8sCluster tears down the test Kind cluster.
func DeleteK8sCluster(clusterName string, kubeconfig string) error {
	cmd := exec.Command("kind", "delete", "cluster",
		"--name", clusterName,
		"--kubeconfig", kubeconfig,
	)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to delete Kind cluster: %v", err)
	}

	return nil
}

// StartServerInK8s deploys the Funnel server in the Kind cluster using Helm.
func StartServerInK8s(clusterName string, chartPath string) error {
	cmd := exec.Command("helm", "upgrade", "--install", "funnel", chartPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy Funnel with Helm: %v", err)
	}

	return nil
}

// TestHelloWorld runs a simple hello world task in the Kubernetes cluster.
func TestHelloWorld(t *testing.T) {
	id := fun.Run(`
    --sh 'echo hello world'
  `)
	task := fun.Wait(id)

	if task.State != tes.State_COMPLETE {
		t.Fatal("expected task to be in complete state; got:", task.State.String())
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout")
	}
}
