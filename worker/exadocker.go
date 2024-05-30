package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Exadocker struct {
	ContainerConfig
}

type ExadockerInfo struct {
	Id    string
	Image string
	Name  string
}

// Run runs the Docker command and blocks until done.
func (exadocker Exadocker) Run(ctx context.Context) error {
	// Sync docker API version info.
	// err := SyncDockerAPIVersion()
	// if err != nil {
	// 	exadocker.Event.Error("failed to sync docker client API version", err)
	// }

	commandArgs := append(exadocker.Driver[1:], "pull", exadocker.Image)
	pullcmd := exec.Command(exadocker.Driver[0], commandArgs...)
	// fmt.Println("DEBUG: exadocker.Driver:", exadocker.Driver)
	// fmt.Println("DEBUG: pullcmd:", pullcmd)

	// Run the command and check for errors.
	err := pullcmd.Run()
	if err != nil {
		exadocker.Event.Error("failed to pull docker image", err)
	}

	args := exadocker.Driver[1:]
	args = append(args, "run", "-i", "--read-only")

	if exadocker.RemoveContainer {
		args = append(args, "--rm")
	}

	if exadocker.Env != nil {
		for k, v := range exadocker.Env {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if exadocker.ContainerName != "" {
		args = append(args, "--name", exadocker.ContainerName)
	}

	if exadocker.Workdir != "" {
		args = append(args, "-w", exadocker.Workdir)
	}

	fmt.Println("DEBUG: exadocker.Volumes:", exadocker.Volumes)
	for _, vol := range exadocker.Volumes {
		fmt.Println("DEBUG: vol:", vol)
		arg := formatExaVolumeArg(vol)
		args = append(args, "-v", arg)
	}

	args = append(args, exadocker.GetImage())
	args = append(args, exadocker.Command...)

	// Roughly: `docker run --rm -i --read-only -w [workdir] -v [bindings] [imageName] [cmd]`
	fmt.Println("DEBUG: exadocker:", exadocker)
	cmd := exec.Command(exadocker.Driver[0], args...)
	fmt.Println("DEBUG: cmd:", cmd)

	if exadocker.Stdin != nil {
		cmd.Stdin = exadocker.Stdin
	}
	fmt.Println("DEBUG: exadocker.Stdout:", exadocker.Stdout)
	if exadocker.Stdout != nil {
		cmd.Stdout = exadocker.Stdout
		// cmd.Stdout = os.Stdout
		fmt.Println("DEBUG: cmd.Stdout:", cmd.Stdout)
	}
	fmt.Println("DEBUG: exadocker.Stderr:", exadocker.Stderr)
	if exadocker.Stderr != nil {
		cmd.Stderr = exadocker.Stderr
		fmt.Println("DEBUG: cmd.Stderr:", cmd.Stderr)
	}
	go exadocker.InspectContainer(ctx)

	// fmt.Println("DEBUG: exadocker.Stdout:", exadocker.Stdout)
	out := cmd.Run()
	exadocker.Event.Info("Command %s Complete exit=%s", strings.Join(args, " "), out)
	fmt.Println("DEBUG: out:", out)
	return out
}

// Stop stops the container.
func (exadocker Exadocker) Stop() error {
	exadocker.Event.Info("Stopping container", "container", exadocker.ContainerName)
	// cmd := exec.Command("docker", "stop", exa.ContainerName)
	driverArgs := strings.Join(exadocker.Driver[1:], " ")
	cmd := exec.Command(exadocker.Driver[0], driverArgs, "rm", "-f", exadocker.ContainerName) //switching to this to be a bit more forceful
	return cmd.Run()
}

func formatExaVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	mode := "rw"
	if v.Readonly {
		mode = "ro"
	}
	// fmt.Println("DEBUG: v.HostPath a:", v.HostPath)
	// v.HostPath = "/mnt/scratch/${SLURM_JOB_ID}" + v.HostPath
	// fmt.Println("DEBUG: v.HostPath b:", v.HostPath)
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, mode)
}

func (exadocker Exadocker) GetImage() string {
	return exadocker.Image
}

func (exadocker Exadocker) GetIO() (io.Reader, io.Writer, io.Writer) {
	return exadocker.Stdin, exadocker.Stdout, exadocker.Stderr
}

func (exadocker *Exadocker) SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	fmt.Println("DEBUG: In SetIO...")
	fmt.Println("DEBUG: exadocker.go stdin:", stdin)
	fmt.Println("DEBUG: exadocker.go stdout:", stdout)
	fmt.Println("DEBUG: exadocker.go stderr:", stderr)
	if stdin != nil {
		exadocker.Stdin = stdin
	}
	if stdout != nil {
		exadocker.Stdout = stdout
	}
	if stderr != nil {
		exadocker.Stderr = stderr
	}
	fmt.Println("DEBUG: exadocker.Stdin:", exadocker.Stdin)
	fmt.Println("DEBUG: exadocker.Stdout:", exadocker.Stdout)
	fmt.Println("DEBUG: exadocker.Stderr:", exadocker.Stderr)
}

func (exadocker Exadocker) Inspect(ctx context.Context) (ContainerConfig, error) {
	info := ContainerConfig{
		Id:    "1234",
		Image: "image",
		Name:  "container",
	}
	return info, nil
}

// inspectContainer inspects the docker container for metadata.
func (exadocker *Exadocker) InspectContainer(ctx context.Context) ContainerConfig {
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
			cmd := exec.CommandContext(ctx, "docker", "inspect", exadocker.ContainerName)
			out, err := cmd.Output()
			if err == nil {
				meta := []ContainerConfig{}
				err := json.Unmarshal(out, &meta)
				if err == nil && len(meta) == 1 {
					exadocker.Event.Info("container metadata",
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
func (exadocker *Exadocker) SyncAPIVersion() error {
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
