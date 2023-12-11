package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DockerCommand is responsible for configuring and running a docker container.
type DockerCommand struct {
	ContainerName   string
	RemoveContainer bool
	Command 	
}

// Run runs the Docker command and blocks until done.
func (dcmd DockerCommand) Run(ctx context.Context) error {
	// Sync docker API version info.
	err := SyncDockerAPIVersion()
	if err != nil {
		dcmd.Event.Error("failed to sync docker client API version", err)
	}

	pullcmd := exec.Command("docker", "pull", dcmd.Image)
	err = pullcmd.Run()
	if err != nil {
		dcmd.Event.Error("failed to pull docker image", err)
	}

	args := []string{"run", "-i", "--read-only"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Env != nil {
		for k, v := range dcmd.Env {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if dcmd.ContainerName != "" {
		args = append(args, "--name", dcmd.ContainerName)
	}

	if dcmd.Workdir != "" {
		args = append(args, "-w", dcmd.Workdir)
	}

	for _, vol := range dcmd.Volumes {
		arg := formatVolumeArg(vol)
		args = append(args, "-v", arg)
	}

	args = append(args, dcmd.Image)
	args = append(args, dcmd.ShellCommand...)

	// Roughly: `docker run --rm -i --read-only -w [workdir] -v [bindings] [imageName] [cmd]`
	dcmd.Event.Info("Running command", "cmd", "docker "+strings.Join(args, " "))
	cmd := exec.Command("docker", args...)

	if dcmd.Stdin != nil {
		cmd.Stdin = dcmd.Stdin
	}
	if dcmd.Stdout != nil {
		cmd.Stdout = dcmd.Stdout
	}
	if dcmd.Stderr != nil {
		cmd.Stderr = dcmd.Stderr
	}
	go dcmd.inspectContainer(ctx)
	out := cmd.Run()
	dcmd.Event.Info("Command %s Complete exit=%s", strings.Join(args, " "), out)
	return out
}

// Stop stops the container.
func (dcmd DockerCommand) Stop() error {
	dcmd.Event.Info("Stopping container", "container", dcmd.ContainerName)
	// cmd := exec.Command("docker", "stop", dcmd.ContainerName)
	cmd := exec.Command("docker", "rm", "-f", dcmd.ContainerName) //switching to this to be a bit more forceful
	return cmd.Run()
}

func formatVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	mode := "rw"
	if v.Readonly {
		mode = "ro"
	}
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, mode)
}

type metadata struct {
	ID    string
	Name  string
	Image string
}

// inspectContainer inspects the docker container for metadata.
func (dcmd *DockerCommand) inspectContainer(ctx context.Context) {
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
			cmd := exec.CommandContext(ctx, "docker", "inspect", dcmd.ContainerName)
			out, err := cmd.Output()
			if err == nil {
				meta := []metadata{}
				err := json.Unmarshal(out, &meta)
				if err == nil && len(meta) == 1 {
					dcmd.Event.Info("container metadata",
						"containerID", meta[0].ID,
						"containerName", meta[0].Name,
						"containerImageHash", meta[0].Image)
					return
				}
			}
		}
	}
}

type dockerVersion struct {
	Client string
	Server string
}

// SyncDockerAPIVersion ensures that the client uses the same API version as
// the server.
func SyncDockerAPIVersion() error {
	if os.Getenv("DOCKER_API_VERSION") == "" {
		cmd := exec.Command("docker", "version", "--format", `{"Server": "{{.Server.APIVersion}}", "Client": "{{.Client.APIVersion}}"}`)
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("docker version command failed: %v", err)
		}
		version := &dockerVersion{}
		err = json.Unmarshal(out, version)
		if err != nil {
			return fmt.Errorf("failed to unmarshal docker version: %v", err)
		}
		if version.Client != version.Server {
			os.Setenv("DOCKER_API_VERSION", version.Server)
		}
	}
	return nil
}
