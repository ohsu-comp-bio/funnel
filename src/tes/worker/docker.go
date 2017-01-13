package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type DockerCmd struct {
	ImageName       string
	Cmd             []string
	Volumes         []Volume
	Workdir         string
	Port            string
	ContainerName   string
	RemoveContainer bool
	Stdin           *os.File
	Stdout          *os.File
	Stderr          *os.File
}

// GetVolumes takes a jobID and returns an array of string.
func formatVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, v.Mode)
}

func (dcmd DockerCmd) Start() (*exec.Cmd, error) {
	log.Printf("Docker Volumes: %s", dcmd.Volumes)

	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Port != "" {
		args = append(args, "-p", dcmd.Port)
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

	args = append(args, dcmd.ImageName)
	args = append(args, dcmd.Cmd...)

	log.Printf("Running command: docker %s", strings.Join(args, " "))
	// Roughly: `docker run --rm -i -w [workdir] -v [bindings] [imageName] [cmd]`
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

	return cmd, cmd.Start()
}

type DockerEngine struct {
	client *client.Client
}

func SetupDockerClient() *DockerEngine {
	client, err := client.NewEnvClient()
	if err != nil {
		log.Printf("Docker Error: %v", err)
		return nil    
	}

	if os.Getenv("DOCKER_API_VERSION") == "" {
		_, err := client.ServerVersion(context.Background())
		if err != nil {
			re := regexp.MustCompile(`([0-9\.]+)`)
			version := re.FindAllString(err.Error(), -1)
			// Error message example: 
			//   Error getting metadata for container: Error response from daemon: 
			//   client is newer than server (client API version: 1.26, server API version: 1.24)
			log.Printf("DOCKER_API_VERSION: %s", version[1])
			os.Setenv("DOCKER_API_VERSION", version[1])
		}
	}
	return &DockerEngine{client: client}
}
