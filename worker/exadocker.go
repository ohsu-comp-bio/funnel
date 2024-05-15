package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/ohsu-comp-bio/funnel/events"
)

type Exadocker struct {
	Image           string
	Command         []string
	Volumes         []Volume
	Workdir         string
	ContainerName   string
	RemoveContainer bool
	Env             map[string]string
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	Event           *events.ExecutorWriter
}

type ExadockerInfo struct {
	Id    string
	Image string
	Name  string
}

var _ ContainerEngine = Exadocker{}

// Run runs the Docker command and blocks until done.
func (exa Exadocker) Run(ctx context.Context) error {
	// Sync docker API version info.
	err := SyncDockerAPIVersion()
	if err != nil {
		exa.Event.Error("failed to sync docker client API version", err)
	}

	pullcmd := exec.Command("docker", "pull", exa.Image)
	err = pullcmd.Run()
	if err != nil {
		exa.Event.Error("failed to pull docker image", err)
	}

	args := []string{"run", "-i", "--read-only"}

	if exa.RemoveContainer {
		args = append(args, "--rm")
	}

	if exa.Env != nil {
		for k, v := range exa.Env {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if exa.ContainerName != "" {
		args = append(args, "--name", exa.ContainerName)
	}

	if exa.Workdir != "" {
		args = append(args, "-w", exa.Workdir)
	}

	for _, vol := range exa.Volumes {
		arg := formatVolumeArg(vol)
		args = append(args, "-v", arg)
	}

	args = append(args, exa.Image)
	args = append(args, exa.Command...)

	// Roughly: `docker run --rm -i --read-only -w [workdir] -v [bindings] [imageName] [cmd]`
	exa.Event.Info("Running command", "cmd", "docker "+strings.Join(args, " "))
	cmd := exec.Command("docker", args...)

	if exa.Stdin != nil {
		cmd.Stdin = exa.Stdin
	}
	if exa.Stdout != nil {
		cmd.Stdout = exa.Stdout
	}
	if exa.Stderr != nil {
		cmd.Stderr = exa.Stderr
	}
	go exa.inspectContainer(ctx)
	out := cmd.Run()
	exa.Event.Info("Command %s Complete exit=%s", strings.Join(args, " "), out)
	return out
}

// Stop stops the container.
func (exa Exadocker) Stop() error {
	exa.Event.Info("Stopping container", "container", exa.ContainerName)
	// cmd := exec.Command("docker", "stop", exa.ContainerName)
	cmd := exec.Command("docker", "rm", "-f", exa.ContainerName) //switching to this to be a bit more forceful
	return cmd.Run()
}

func (exa Exadocker) Inspect(ctx context.Context) (ContainerInfo, error) {
	info := ContainerInfo{
		Id:    "1234",
		Image: "image",
		Name:  "container",
	}
	return info, nil
}

// inspectContainer inspects the docker container for metadata.
func (exa *Exadocker) inspectContainer(ctx context.Context) {
	// Give the container time to start.
	time.Sleep(2 * time.Second)

	// Inspect the container for metadata
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cmd := exec.CommandContext(ctx, "docker", "inspect", exa.ContainerName)
			out, err := cmd.Output()
			if err == nil {
				meta := []metadata{}
				err := json.Unmarshal(out, &meta)
				if err == nil && len(meta) == 1 {
					exa.Event.Info("container metadata",
						"containerID", meta[0].ID,
						"containerName", meta[0].Name,
						"containerImageHash", meta[0].Image)
					return
				}
			}
		}
	}
}
