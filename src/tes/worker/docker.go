package tesTaskEngineWorker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	pbe "tes/ga4gh"
	"time"
)

// DockerCmd is responsible for configuring and running a docker container.
type DockerCmd struct {
	ImageName       string
	CmdString       []string
	Volumes         []Volume
	Workdir         string
	PortBindings    []*pbe.PortMapping
	ContainerName   string
	RemoveContainer bool
	Stdin           *os.File
	Stdout          *os.File
	Stderr          *os.File
	Cmd             *exec.Cmd
	// store last 200 lines of both stdout and stderr
	Log map[string][]byte
}

// GetVolumes takes a jobID and returns an array of string.
func formatVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, v.Mode)
}

// SetupCommand sets up the command to be run and sets DockerCmd.Cmd.
// Essentially it prepares commandline arguments for Docker.
func (dcmd DockerCmd) SetupCommand() (*DockerCmd, error) {
	log.Printf("Docker Volumes: %s", dcmd.Volumes)

	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.PortBindings != nil {
		log.Printf("Docker Port Bindings: %v", dcmd.PortBindings)
		for i := range dcmd.PortBindings {
			hostPortNum := int(dcmd.PortBindings[i].HostBinding)
			if hostPortNum <= 1024 && hostPortNum != 0 {
				return nil, fmt.Errorf("Error cannot use restricted ports")
			}
			hostPort := strconv.Itoa(hostPortNum)
			containerPort := strconv.Itoa(int(dcmd.PortBindings[i].ContainerPort))
			args = append(args, "-p", hostPort+":"+containerPort)
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
		return nil, fmt.Errorf("Error creating StdoutPipe for Cmd: %s", err)
	}

	stdoutScanner := bufio.NewScanner(stdoutReader)
	go func() {
		for stdoutScanner.Scan() {
			s := stdoutScanner.Text()
			dcmd.Stdout.WriteString(s + "\n")
			dcmd.Log["Stdout"] = updateAndTrim(dcmd.Log["Stdout"], []byte(s+"\n"))
		}
	}()

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Error creating StderrPipe for Cmd: %s", err)
	}

	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		for stderrScanner.Scan() {
			e := stderrScanner.Text()
			dcmd.Stderr.WriteString(e + "\n")
			dcmd.Log["Stderr"] = updateAndTrim(dcmd.Log["Stderr"], []byte(e+"\n"))
		}
	}()

	return &dcmd, nil
}

func updateAndTrim(l []byte, v []byte) []byte {
	// 10 KB max stored
	max := 10000
	l = append(l[:], v[:]...)
	if len(l) > max {
		return l[len(l)-max : len(l)]
	}
	return l
}

// InspectContainer returns metadata about the container (calls "docker inspect").
func (dcmd DockerCmd) InspectContainer() (*types.ContainerJSON, error) {
	log.Printf("Fetching container metadata")
	dclient := setupDockerClient()
	// close the docker client connection
	defer dclient.Close()
	// Set timeout
	timeout := time.After(time.Second * 10)

	for {
		metadata, err := dclient.ContainerInspect(context.Background(), dcmd.ContainerName)
		select {
		case <-timeout:
			return nil, fmt.Errorf("Error getting metadata for container: %s", err)
		default:
			switch {
			case err == nil && metadata.State.Running == true:
				return &metadata, err
			}
		}
	}
}

// StopContainer stops the container.
func (dcmd DockerCmd) StopContainer() error {
	log.Printf("Stopping container %s", dcmd.ContainerName)
	dclient := setupDockerClient()
	// close the docker client connection
	defer dclient.Close()
	// Set timeout
	timeout := time.Second * 10
	// Issue stop call
	err := dclient.ContainerStop(context.Background(), dcmd.ContainerName, &timeout)
	return err
}

func setupDockerClient() *client.Client {
	dclient, err := client.NewEnvClient()
	if err != nil {
		log.Printf("Docker Error: %v", err)
		return nil
	}

	// If the api version is not set test if the client can communicate with the
	// server; if not infer API version from error message and inform the client
	// to use that version for future communication
	if os.Getenv("DOCKER_API_VERSION") == "" {
		_, err := dclient.ServerVersion(context.Background())
		if err != nil {
			re := regexp.MustCompile(`([0-9\.]+)`)
			version := re.FindAllString(err.Error(), -1)
			// Error message example:
			//   Error getting metadata for container: Error response from daemon: client is newer than server (client API version: 1.26, server API version: 1.24)
			log.Printf("DOCKER_API_VERSION: %s", version[1])
			os.Setenv("DOCKER_API_VERSION", version[1])
			return setupDockerClient()
		}
	}
	return dclient
}
