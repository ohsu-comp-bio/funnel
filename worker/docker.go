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
func (docker Docker) Run(ctx context.Context) error {
	// Sync docker API version info.
	err := docker.SyncAPIVersion()
	if err != nil {
		docker.Event.Error("failed to sync docker client API version", err)
	}

	pullcmd := exec.Command("docker", "pull", docker.Image)
	err = pullcmd.Run()
	if err != nil {
		docker.Event.Error("failed to pull docker image", err)
	}

	var args []string

	if len(docker.ContainerConfig.DriverCommand) > 1 {
		// Merge driver parts and command parts
		args = append(args, docker.ContainerConfig.DriverCommand[1:]...)
	}

	args = append(args, "run", "-i", "--read-only")

	if docker.RemoveContainer {
		args = append(args, "--rm")
	}

	if docker.Env != nil {
		for k, v := range docker.Env {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if docker.Name != "" {
		args = append(args, "--name", docker.Name)
	}

	if docker.Workdir != "" {
		args = append(args, "-w", docker.Workdir)
	}

	for _, vol := range docker.Volumes {
		arg := formatVolumeArg(vol)
		args = append(args, "-v", arg)
	}

	args = append(args, docker.Image)
	args = append(args, docker.Command...)

	// Roughly: `docker run --rm -i --read-only -w [workdir] -v [bindings] [imageName] [cmd]`
	docker.Event.Info("Running command", "cmd", docker.ContainerConfig.DriverCommand[0]+" "+strings.Join(args, " "))
	cmd := exec.Command(docker.ContainerConfig.DriverCommand[0], args...)

	if docker.Stdin != nil {
		cmd.Stdin = docker.Stdin
	}
	if docker.Stdout != nil {
		cmd.Stdout = docker.Stdout
	}
	if docker.Stderr != nil {
		cmd.Stderr = docker.Stderr
	}
	go docker.InspectContainer(ctx)
	out := cmd.Run()
	docker.Event.Info("Command %s Complete exit=%s", strings.Join(args, " "), out)
	return out
}

// Stop stops the container.
func (docker Docker) Stop() error {
	docker.Event.Info("Stopping container", "container", docker.Name)
	// cmd := exec.Command("docker", "stop", docker.Name)
	cmd := exec.Command("docker", "rm", "-f", docker.Name) //switching to this to be a bit more forceful
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

func (docker Docker) GetImage() string {
	return docker.Image
}

func (docker Docker) GetIO() (io.Reader, io.Writer, io.Writer) {
	return docker.Stdin, docker.Stdout, docker.Stderr
}

func (docker *Docker) SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	if stdin != nil && stdin != (*os.File)(nil) {
		docker.Stdin = stdin
	}
	if stdout != nil && stdout != (*os.File)(nil) {
		docker.Stdout = stdout
	}
	if stderr != nil && stderr != (*os.File)(nil) {
		docker.Stderr = stderr
	}
}

// inspectContainer inspects the docker container for metadata.
func (docker *Docker) InspectContainer(ctx context.Context) ContainerConfig {
	// Give the container time to start.
	time.Sleep(2 * time.Second)

	// Inspect the container for metadata
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			return ContainerConfig{}
		case <-ticker.C:
			cmd := exec.CommandContext(ctx, "docker", "inspect", docker.Name)
			out, err := cmd.Output()
			if err == nil {
				fmt.Println("DEBUG: string(out):", string(out))

				meta := []ContainerConfig{}
				err = json.Unmarshal(out, &meta)
				if err == nil && len(meta) == 1 {
					docker.Event.Info("container metadata",
						"containerID", meta[0].Id,
						"containerName", meta[0].Name,
						"containerImageHash", meta[0].Image)
					return meta[0]
				}
			}
		}
	}

	return ContainerConfig{}
}

// SyncDockerAPIVersion ensures that the client uses the same API version as
// the server.
func (docker *Docker) SyncAPIVersion() error {
	if os.Getenv("DOCKER_API_VERSION") == "" {
		cmd := exec.Command("docker", "version", "--format", `{"Server": "{{.Server.APIVersion}}", "Client": "{{.Client.APIVersion}}"}`)
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("docker version command failed: %v", err)
		}
		version := &ContainerVersion{}
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
