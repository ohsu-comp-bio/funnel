package tesTaskEngineWorker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	pbe "tes/ga4gh"
	"time"
)

// DockerCmd is responsible for configuring and running a docker container.
type DockerCmd struct {
	ImageName       string
	CmdString       []string
	Volumes         []Volume
	Workdir         string
	Ports           []*pbe.Ports
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
	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Ports != nil {
		for i := range dcmd.Ports {
			hostPort := dcmd.Ports[i].Host
			containerPort := dcmd.Ports[i].Container
			if hostPort <= 1024 && hostPort != 0 {
				return nil, fmt.Errorf("Error cannot use restricted ports")
			}
			args = append(args, "-p", fmt.Sprintf("%d:%d", hostPort, containerPort))
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

	log.Debug("DockerCmd", "dmcd", dcmd)

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
		return l[len(l)-max:]
	}
	return l
}

// InspectContainer returns metadata about the container (calls "docker inspect").
func (dcmd DockerCmd) InspectContainer(ctx context.Context) []*pbe.Ports {
	log.Info("Fetching container metadata")
	dclient := setupDockerClient()
	// close the docker client connection
	defer dclient.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			metadata, err := dclient.ContainerInspect(ctx, dcmd.ContainerName)
			if err == nil && metadata.State.Running == true {
				var portMap []*pbe.Ports
				// extract exposed host port from
				// https://godoc.org/github.com/docker/go-connections/nat#PortMap
				for k, v := range metadata.NetworkSettings.Ports {
					// will end up taking the last binding listed
					for i := range v {
						p := strings.Split(string(k), "/")
						containerPort, err := strconv.Atoi(p[0])
						//TODO handle errors
						if err != nil {
							return nil
						}
						hostPort, err := strconv.Atoi(v[i].HostPort)
						//TODO handle errors
						if err != nil {
							return nil
						}
						portMap = append(portMap, &pbe.Ports{
							Container: int32(containerPort),
							Host:      int32(hostPort),
						})
					}
				}
				return portMap
			}
		}
	}
}

// StopContainer stops the container.
func (dcmd DockerCmd) StopContainer() error {
	log.Info("Stopping container", "container", dcmd.ContainerName)
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
		log.Info("Docker error", "err", err)
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
			log.Debug("DOCKER_API_VERSION", "version", version[1])
			os.Setenv("DOCKER_API_VERSION", version[1])
			return setupDockerClient()
		}
	}
	return dclient
}
