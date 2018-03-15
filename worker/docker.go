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
	"github.com/ohsu-comp-bio/funnel/util/dockerutil"
)

// DockerCommand is responsible for configuring and running a docker container.
type DockerCommand struct {
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

// Run runs the Docker command and blocks until done.
func (dcmd DockerCommand) Run(ctx context.Context) error {
	// (Hopefully) temporary hack to sync docker API version info.
	// Don't need the client here, just the logic inside NewDockerClient().
	_, derr := dockerutil.NewDockerClient()
	if derr != nil {
		dcmd.Event.Error("Can't connect to Docker", derr)
		return derr
	}

	pullcmd := exec.Command("docker", "pull", dcmd.Image)
	pullcmd.Run()

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
	args = append(args, dcmd.Command...)

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
	return cmd.Run()
}

// Stop stops the container.
func (dcmd DockerCommand) Stop() error {
	dcmd.Event.Info("Stopping container", "container", dcmd.ContainerName)
	dclient, derr := dockerutil.NewDockerClient()
	if derr != nil {
		return derr
	}
	// close the docker client connection
	defer dclient.Close()
	// Set timeout
	timeout := time.Second * 10
	// Issue stop call
	// TODO is context.Background right?
	err := dclient.ContainerStop(context.Background(), dcmd.ContainerName, &timeout)
	return err
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
