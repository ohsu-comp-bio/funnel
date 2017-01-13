package tesTaskEngineWorker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
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
	ExecCmd         *exec.Cmd
}

// GetVolumes takes a jobID and returns an array of string.
func formatVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, v.Mode)
}

func (dcmd DockerCmd) SetupCommand() *DockerCmd {
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
	dcmd.ExecCmd = cmd

	if dcmd.Stdin != nil {
		cmd.Stdin = dcmd.Stdin
	}
	if dcmd.Stdout != nil {
		cmd.Stdout = dcmd.Stdout
	}
	if dcmd.Stderr != nil {
		cmd.Stderr = dcmd.Stderr
	}

	return &dcmd
}

func (dcmd DockerCmd) GetContainerMetadata() (string, error) {
	log.Printf("Fetching container metadata")
	deng := SetupDockerClient()

	// Set timeout
	timeout := time.After(time.Second * 10)

	for {
		meta, err := deng.client.ContainerInspect(context.Background(), dcmd.ContainerName)
		select {
		case <- timeout:
			return "", fmt.Errorf("Error getting metadata for container: %s", err)
		default:
			switch {
			case err == nil:
			// close the docker client connection
			deng.client.Close()

			// TODO congifure which fields to keep from docker inspect
			// whitelist := []string
			// for k, v := range meta {
			//  if k in not in whitelist {
			//    delete(meta, k)
			// }

			metadata, _ := json.Marshal(meta)
			return string(metadata), err
			}
		}
	}
}

func (dcmd DockerCmd) StopContainer() (error) {
	log.Printf("Stoping container %s", dcmd.ContainerName)
	deng := SetupDockerClient()	
	// Set timeout
	timeout := time.Second * 10
	// Issue stop call
	err := deng.client.ContainerStop(context.Background(), dcmd.ContainerName, &timeout)
	// close the docker client connection
	deng.client.Close()
	return err
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
			return SetupDockerClient()
		}
	}
	return &DockerEngine{client: client}
}
