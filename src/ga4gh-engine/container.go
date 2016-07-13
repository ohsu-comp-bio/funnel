
package ga4gh_taskengine

import (
	"os"
	"log"
	"strings"
	"github.com/fsouza/go-dockerclient"
)


type ContainerManager interface {
	Run(container string, args []string, binds[] string, remove bool, stdout_path *string, stderr_path *string) error
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
	return &DockerDirect{ client:client }
}

func (self *DockerDirect) Run(containerName string, args []string, binds[] string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	create_config := docker.Config{
		Image:containerName,
		Cmd:args,
		AttachStderr:true,
		AttachStdout:true,
	}

	if _, ok := self.client.InspectImage(containerName); ok != nil {
		log.Printf("Image %s not found", containerName)
		tmp := strings.Split(containerName, ":")
		pull_opt := docker.PullImageOptions{ Repository: tmp[0], Tag: tmp[1] }
		if ok := self.client.PullImage( pull_opt, docker.AuthConfiguration{} ); ok != nil {
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
	err = self.client.StartContainer(container.ID, &docker.HostConfig {
		Binds: binds,
	})

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return 0, err
	}

	log.Printf("Attaching Container: %s", container.ID)
	exit_code, err := self.client.WaitContainer(container.ID)

	logOpts := docker.LogsOptions{Container:container.ID, Stdout:false, Stderr:false}

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
	self.client.RemoveContainer(docker.RemoveContainerOptions{ID:container.ID,RemoveVolumes:true})
	return exit_code, nil
}