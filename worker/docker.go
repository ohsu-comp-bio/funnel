package worker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DockerCmd is responsible for configuring and running a docker container.
type DockerCmd struct {
	ImageName       string
	Cmd             []string
	Volumes         []Volume
	Workdir         string
	Ports           []*tes.Ports
	ContainerName   string
	RemoveContainer bool
	Environ         map[string]string
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
}

// Run runs the Docker command and blocks until done.
func (dcmd DockerCmd) Run() error {
	// (Hopefully) temporary hack to sync docker API version info.
	// Don't need the client here, just the logic inside NewDockerClient().
	_, derr := util.NewDockerClient()
	if derr != nil {
		return fmt.Errorf("Can't connect to Docker: %s", derr)
	}

	cmd := exec.Command("docker", dcmd.Args()...)

	if dcmd.Stdin != nil {
		cmd.Stdin = dcmd.Stdin
	}
	if dcmd.Stdout != nil {
		cmd.Stdout = dcmd.Stdout
	}
	if dcmd.Stderr != nil {
		cmd.Stderr = dcmd.Stderr
	}
	return cmd.Run()
}

// Args returns the arguments to the "docker" command.
// Roughly: run --rm -i -w [workdir] -v [bindings] [imageName] [cmd]
func (dcmd DockerCmd) Args() []string {
	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Environ != nil {
		for k, v := range dcmd.Environ {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if dcmd.Ports != nil {
		for i := range dcmd.Ports {
			hostPort := dcmd.Ports[i].Host
			containerPort := dcmd.Ports[i].Container
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
	return args
}

// Inspect returns metadata about the container (calls "docker inspect").
func (dcmd DockerCmd) Inspect(ctx context.Context) ([]*tes.Ports, error) {
	dclient, derr := util.NewDockerClient()
	if derr != nil {
		return nil, derr
	}
	// close the docker client connection
	defer dclient.Close()
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			metadata, err := dclient.ContainerInspect(ctx, dcmd.ContainerName)
			if client.IsErrContainerNotFound(err) {
				break
			}
			if err != nil {
				break
			}
			if metadata.State.Running == true {
				var portMap []*tes.Ports
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
						portMap = append(portMap, &tes.Ports{
							Container: uint32(containerPort),
							Host:      uint32(hostPort),
						})
					}
				}
				return portMap, nil
			}
		}
	}
}

// Stop stops the container.
func (dcmd DockerCmd) Stop() error {
	dclient, derr := util.NewDockerClient()
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
