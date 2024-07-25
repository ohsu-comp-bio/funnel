package worker

import (
	"bytes"
	"context"
	"testing"
	"time"

	"math/rand"

	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
)

var defaultConfig = ContainerConfig{
	Id:              "123",
	Image:           "alpine",
	Name:            "funnel-test-" + RandomString(6),
	DriverCommand:   "docker",
	Command:         []string{"sh", "-c", "echo Hello, World!"},
	RunCommand:      "run --name {{.Name}} {{.Image}} {{.Command}}",
	PullCommand:     "pull {{.Image}}",
	RemoveContainer: true,
	Event: events.NewExecutorWriter("123", 1, 1, &events.Logger{
		Log: logger.NewLogger("test", logger.DefaultConfig()),
	}),
}

func TestDockerRun(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	err := docker.Run(context.Background())
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestDockerExecuteCommand(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	err := docker.executeCommand(context.Background(), "run --rm alpine echo Hello, World!", true)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestDockerStop(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Run the container command in a separate goroutine
	go func() {
		err := docker.executeCommand(ctx, "run --rm alpine sleep 30", true)
		if err != nil && ctx.Err() == nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	}()

	// Give the container some time to start
	time.Sleep(2 * time.Second)

	// Stop the container
	err := docker.Stop()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Cancel the context to stop the goroutine if it is still running
	cancel()
}

func TestFormatVolumeArg(t *testing.T) {
	volume := Volume{
		HostPath:      "/path/to/source",
		ContainerPath: "/path/to/destination",
		Readonly:      true,
	}
	expected := "/path/to/source:/path/to/destination:ro"
	result := formatVolumeArg(volume)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestDockerGetImage(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	expected := "alpine"
	result := docker.GetImage()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestDockerSetIO(t *testing.T) {
	docker := Docker{}
	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	docker.SetIO(stdin, stdout, stderr)
	if docker.Stdin != stdin || docker.Stdout != stdout || docker.Stderr != stderr {
		t.Errorf("Expected stdin, stdout, and stderr to be set correctly")
	}
}

func TestDockerInspectContainer(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	config := docker.InspectContainer(context.Background())
	if config.Id == "" {
		t.Errorf("Expected non-nil container config")
	}
}

func TestDockerSyncAPIVersion(t *testing.T) {
	docker := Docker{
		ContainerConfig: defaultConfig,
	}
	err := docker.SyncAPIVersion()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
