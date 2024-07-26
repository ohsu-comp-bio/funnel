package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// Docker is responsible for configuring and running a docker container.
type Docker struct {
	ContainerConfig
}

// Run runs the Docker command and blocks until done.
func (docker Docker) Run(ctx context.Context) error {
	// Sync docker API version info.
	err := docker.SyncAPIVersion()
	if err != nil {
		docker.Event.Error("failed to sync docker client API version", err)
	}

	err = docker.executeCommand(ctx, docker.PullCommand, false)
	if err != nil {
		docker.Event.Error("failed to pull docker image", err)
	}

	err = docker.executeCommand(ctx, docker.RunCommand, true)
	if err != nil {
		docker.Event.Error("failed to run docker container", err)
	}

	return err
}

// Stop stops the container.
func (docker Docker) Stop() error {
	docker.Event.Info("Stopping container", "container", docker.Name)
	err := docker.executeCommand(context.Background(), docker.StopCommand, false)
	if err != nil {
		docker.Event.Error("failed to stop docker container", err)
		return err
	}
	return nil
}

func (docker Docker) executeCommand(ctx context.Context, commandTemplate string, enableIO bool) error {
	var usingCommand bool = false
	if strings.Contains(commandTemplate, "{{.Command}}") {
		usingCommand = true
		commandTemplate = strings.ReplaceAll(commandTemplate, "{{.Command}}", "")
	}

	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template for command: %w", err)
	}

	var cmdBuffer bytes.Buffer
	err = tmpl.Execute(&cmdBuffer, docker.ContainerConfig)
	if err != nil {
		return fmt.Errorf("failed to execute template for command: %w", err)
	}

	cmdParts := strings.Fields(cmdBuffer.String())
	if usingCommand {
		go docker.InspectContainer(ctx)
		cmdParts = append(cmdParts, docker.Command...)
	}

	driverCmd := strings.Fields(docker.DriverCommand)
	var cmd *exec.Cmd
	if len(driverCmd) > 1 {
		cmdArgs := append(driverCmd[1:], cmdParts...)
		cmd = exec.CommandContext(ctx, driverCmd[0], cmdArgs...)
	} else {
		cmd = exec.CommandContext(ctx, driverCmd[0], cmdParts...)
	}

	if enableIO {
		if docker.Stdin != nil {
			cmd.Stdin = docker.Stdin
		}
		if docker.Stdout != nil {
			cmd.Stdout = docker.Stdout
		}
		if docker.Stderr != nil {
			cmd.Stderr = docker.Stderr
		}
	}

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
		var args []string
		driverCmd := strings.Fields(docker.ContainerConfig.DriverCommand)

		if len(docker.ContainerConfig.DriverCommand) > 1 {
			// Merge driver parts and command parts
			args = append(args, driverCmd[1:]...)
		}

		args = append(args, "version", "--format", `{"Server": "{{.Server.APIVersion}}", "Client": "{{.Client.APIVersion}}"}`)
		cmd := exec.Command(driverCmd[0], args...)
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
