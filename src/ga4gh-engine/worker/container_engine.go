package ga4gh_taskengine_worker

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"golang.org/x/net/context"
	"io"
	"log"
	"os"
	"strings"
)

type DockerEngine struct {
	client *client.Client
}

func NewDockerEngine() *DockerEngine {
	client, err := client.NewEnvClient()
	if err != nil {
		log.Printf("Docker Error\n")
		return nil
	}
	return &DockerEngine{client: client}
}

func (self *DockerEngine) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	list, err := self.client.ImageList(context.Background(), types.ImageListOptions{MatchName: containerName})

	if err != nil || len(list) == 0 {
		log.Printf("Image %s not found: %s", containerName, err)
		pull_opt := types.ImagePullOptions{}
		r, err := self.client.ImagePull(context.Background(), containerName, pull_opt)
		if err != nil {
			log.Printf("Image not pulled: %s", err)
			return -1, err
		}
		for {
			l := make([]byte, 1000)
			_, e := r.Read(l)
			if e == io.EOF {
				break
			}
			log.Printf("%s", l)
		}
		r.Close()
		log.Printf("Image Pulled")
	}

	container, err := self.client.ContainerCreate(context.Background(),
		&container.Config{Cmd: args, Image: containerName, Tty: true},
		&container.HostConfig{Binds: binds},
		&network.NetworkingConfig{},
		"",
	)

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return 0, err
	}

	log.Printf("Starting Docker (mount: %s): %s", strings.Join(binds, ","), strings.Join(args, " "))
	err = self.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return 0, err
	}

	log.Printf("Attaching Container: %s", container.ID)
	exit_code, err := self.client.ContainerWait(context.Background(), container.ID)
	if err != nil {
		log.Printf("docker %s error: %s", container.ID, err)
	} else {
		log.Printf("docker %s complete", container.ID, err)
	}

	if stdout != nil {
		stdout_log, _ := self.client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{ShowStdout: true, Details: false})
		buffer := make([]byte, 10240)
		for {
			l, e := stdout_log.Read(buffer)
			if e == io.EOF {
				break
			}
			stdout.Write(buffer[:l])
		}
		stdout_log.Close()
	}

	if stderr != nil {
		stderr_log, _ := self.client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{ShowStderr: true})
		buffer := make([]byte, 10240)
		for {
			l, e := stderr_log.Read(buffer)
			if e == io.EOF {
				break
			}
			stderr.Write(buffer[:l])
		}
		stderr_log.Close()
	}
	self.client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	return exit_code, nil
}
