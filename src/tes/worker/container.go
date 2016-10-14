package ga4gh_taskengine_worker

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"strings"
)

type ContainerManager interface {
	Run(container string, args []string, binds []string, workdir string, remove bool, stdout_path *string, stderr_path *string) error
}

type DockerDirect struct {
	client *docker.Client
}

func NewDockerDirect() *DockerDirect {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		log.Printf("Docker Error\n")
		return nil
	}
	return &DockerDirect{client: client}
}

func (self *DockerDirect) Run(containerName string, args []string, binds []string, workdir string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	create_config := docker.Config{
		Image:        containerName,
		Cmd:          args,
		AttachStderr: true,
		AttachStdout: true,
	}

	if len(workdir) > 0 {
		create_config.WorkingDir = workdir
	}

	if _, ok := self.client.InspectImage(containerName); ok != nil {
		log.Printf("Image %s not found", containerName)
		tmp := strings.Split(containerName, ":")
		rep := tmp[0]
		tag := "latest"
		if len(tmp) > 1 {
			tag = tmp[1]
		}
		pull_opt := docker.PullImageOptions{Repository: rep, Tag: tag}
		if ok := self.client.PullImage(pull_opt, docker.AuthConfiguration{}); ok != nil {
			log.Printf("Image not pulled: %s", ok)
			return -1, ok
		}
		log.Printf("Image Pulled")
	}

	container, err := self.client.CreateContainer(docker.CreateContainerOptions{
		Config: &create_config,
	})
	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return 0, err
	}

	log.Printf("Starting Docker (mount: %s): %s", strings.Join(binds, ","), strings.Join(args, " "))
	err = self.client.StartContainer(container.ID, &docker.HostConfig{
		Binds: binds,
	})

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return 0, err
	}

	log.Printf("Attaching Container: %s", container.ID)
	exit_code, err := self.client.WaitContainer(container.ID)
	if err != nil {
		log.Printf("Docker run Error: %s", err)
	}

	logOpts := docker.LogsOptions{Container: container.ID, Stdout: false, Stderr: false}

	if stdout != nil {
		logOpts.Stdout = true
		logOpts.OutputStream = stdout
	}
	if stderr != nil {
		logOpts.Stderr = true
		logOpts.ErrorStream = stderr
	}

	self.client.Logs(logOpts)
	if err != nil {
		log.Printf("docker %s error: %s", container.ID, err)
	} else {
		log.Printf("docker %s complete", container.ID, err)
	}
	self.client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID, RemoveVolumes: true})
	return exit_code, nil
}
