package tesTaskEngineWorker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type DockerCmd struct {
	ImageName       string
	CmdString       []string
	Volumes         []Volume
	Workdir         string
	Port            string
	ContainerName   string
	RemoveContainer bool
	Stdin           *os.File
	Stdout          *os.File
	Stderr          *os.File
	Cmd             *exec.Cmd
	// store last 200 lines of both stdout and stderr
	Log             map[string][]string
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
	args = append(args, dcmd.CmdString...)

	// Roughly: `docker run --rm -i -w [workdir] -v [bindings] [imageName] [cmd]`
	cmd := exec.Command("docker", args...)
	dcmd.Cmd = cmd

	if dcmd.Stdin != nil {
		cmd.Stdin = dcmd.Stdin
	}

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe for Cmd", err)
	}

	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			s := stdoutScanner.Text()
			dcmd.Stdout.WriteString(s + "/n")
			dcmd.Log["Stdout"] = UpdateAndTrim(dcmd.Log["Stdout"], s)
		}
	}()

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating StderrPipe for Cmd", err)
	}

	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			e := stderrScanner.Text()
			log.Printf("Stderr: %s\n", e)
			dcmd.Stderr.WriteString(e + "/n")
			dcmd.Log["Stderr"] = UpdateAndTrim(dcmd.Log["Stderr"], e)
		}
	}()

	return &dcmd
}

func UpdateAndTrim(l []string, v string) []string {
	// TODO does it make sense to limit these logs to 200 lines?
	max := 200
	l = append(l, v)
	if len(l) > max {
		return l[len(l)-max:len(l)]
	}
	return l
}

func (dcmd DockerCmd) InspectContainer() (*types.ContainerJSON, error) {
	log.Printf("Fetching container metadata")
	deng := SetupDockerClient()

	// Set timeout
	timeout := time.After(time.Second * 10)

	for {
		metadata, err := deng.client.ContainerInspect(context.Background(), dcmd.ContainerName)
		select {
		case <-timeout:
			return nil, fmt.Errorf("Error getting metadata for container: %s", err)
		default:
			switch {
			case err == nil:
				// close the docker client connection
				deng.client.Close()
				return &metadata, err
			}
		}
	}
}

func (dcmd DockerCmd) StopContainer() error {
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

	// If the api version is not set test if the client can communicate with the 
	// server; if not infer API version from error message and inform the client 
	// to use that version for future communication
	if os.Getenv("DOCKER_API_VERSION") == "" {
		_, err := client.ServerVersion(context.Background())
		if err != nil {
			re := regexp.MustCompile(`([0-9\.]+)`)
			version := re.FindAllString(err.Error(), -1)
			// Error message example:
			//   Error getting metadata for container: Error response from daemon: client is newer than server (client API version: 1.26, server API version: 1.24)
			log.Printf("DOCKER_API_VERSION: %s", version[1])
			os.Setenv("DOCKER_API_VERSION", version[1])
			return SetupDockerClient()
		}
	}
	return &DockerEngine{client: client}
}
