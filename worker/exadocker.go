package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	for _, vol := range exadocker.Volumes {
		arg := formatExaVolumeArg(vol)
		args = append(args, "-v", arg)
	}

	args = append(args, exadocker.GetImage())
	args = append(args, exadocker.Command...)

	// Roughly: `docker run --rm -i --read-only -w [workdir] -v [bindings] [imageName] [cmd]`
	cmd := exec.Command(exadocker.Driver[0], args...)
	// fmt.Println("DEBUG: cmd:", cmd)

	if exadocker.Stdin != nil {
		cmd.Stdin = exadocker.Stdin
	}
	if exadocker.Stdout != nil {
		fmt.Println("DEBUG: exadocker.Stdout:", exadocker.Stdout)
		cmd.Stdout = exadocker.Stdout
		fmt.Println("DEBUG: cmd.Stdout:", cmd.Stdout)
	}
	if exadocker.Stderr != nil {
		cmd.Stderr = exadocker.Stderr
	}
	go exadocker.inspectContainer(ctx)
	out := cmd.Run()
	exadocker.Event.Info("Command %s Complete exit=%s", strings.Join(args, " "), out)
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
	v.HostPath = "/mnt/scratch/${SLURM_JOB_ID}" + v.HostPath
	// fmt.Println("DEBUG: v.HostPath b:", v.HostPath)
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, mode)
}

func (exadocker Exadocker) GetImage() string {
	return exadocker.Image
}

func (exadocker Exadocker) GetIO() (io.Reader, io.Writer, io.Writer) {
	return exadocker.Stdin, exadocker.Stdout, exadocker.Stderr
}

func (exadocker Exadocker) SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	if stdin != nil {
		exadocker.Stdin = stdin
	}
	if stdout != nil {
		exadocker.Stdout = stdout
	}
	if stderr != nil {
		exadocker.Stderr = stderr
	}
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
func (exadocker *Exadocker) inspectContainer(ctx context.Context) {
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
			driverArgs := strings.Join(exadocker.Driver[1:], " ")
			cmd := exec.CommandContext(ctx, exadocker.Command[0], driverArgs, "inspect", exadocker.ContainerName)
			out, err := cmd.Output()
			if err == nil {
				meta := []metadata{}
				err := json.Unmarshal(out, &meta)
				if err == nil && len(meta) == 1 {
					exadocker.Event.Info("container metadata",
						"containerID", meta[0].ID,
						"containerName", meta[0].Name,
						"containerImageHash", meta[0].Image)
					return
				}
			}
		}
	}
}
