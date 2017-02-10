package worker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io"
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
	Cmd             []string
	Volumes         []Volume
	Workdir         string
	Ports           []*pbe.Ports
	ContainerName   string
	RemoveContainer bool
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
}

// SetupCommand sets up the command to be run and sets DockerCmd.Cmd.
// Essentially it prepares commandline arguments for Docker.
func (dcmd DockerCmd) Run() error {
	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Ports != nil {
		for i := range dcmd.Ports {
			hostPort := dcmd.Ports[i].Host
			containerPort := dcmd.Ports[i].Container
			// TODO move to validation?
			if hostPort <= 1024 && hostPort != 0 {
				return fmt.Errorf("Error cannot use restricted ports")
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
	args = append(args, dcmd.Cmd...)

	log.Debug("DockerCmd", "dmcd", dcmd)

	// Roughly: `docker run --rm -i -w [workdir] -v [bindings] [imageName] [cmd]`
	log.Info("Running command", "cmd", "docker "+strings.Join(args, " "))
	cmd := exec.Command("docker", args...)

	if dcmd.Stdin != nil {
		cmd.Stdin = dcmd.Stdin
	}
	return cmd.Run()
}

// Inspect returns metadata about the container (calls "docker inspect").
func (dcmd DockerCmd) Inspect(ctx context.Context) ([]*pbe.Ports, error) {
	log.Info("Fetching container metadata")
	dclient := setupDockerClient()
	// close the docker client connection
	defer dclient.Close()
	for {
		select {
		case <-ctx.Done():
			return nil, nil
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
						if err != nil {
							return nil, err
						}
						hostPort, err := strconv.Atoi(v[i].HostPort)
						if err != nil {
							return nil, err
						}
						portMap = append(portMap, &pbe.Ports{
							Container: int32(containerPort),
							Host:      int32(hostPort),
						})
					}
				}
				return portMap, nil
			}
		}
	}
}

// StopContainer stops the container.
func (dcmd DockerCmd) Stop() error {
	log.Info("Stopping container", "container", dcmd.ContainerName)
	dclient := setupDockerClient()
	// close the docker client connection
	defer dclient.Close()
	// Set timeout
	timeout := time.Second * 10
	// Issue stop call
	// TODO is context.Background right?
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

// GetVolumes takes a jobID and returns an array of string.
func formatVolumeArg(v Volume) string {
	// `o` is structed as "HostPath:ContainerPath:Mode".
	return fmt.Sprintf("%s:%s:%s", v.HostPath, v.ContainerPath, v.Mode)
}
