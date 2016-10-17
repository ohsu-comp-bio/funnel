package tes_taskengine_worker

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func NewDockerEngine() *DockerCmd {
	return &DockerCmd{}
}

type DockerCmd struct {
}

func (self DockerCmd) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	log.Printf("Docker Binds: %s", binds)

	docker_args := []string{"run", "--rm", "-i"}

	if workdir != "" {
		docker_args = append(docker_args, "-w", workdir)
	}

	for _, i := range binds {
		docker_args = append(docker_args, "-v", i)
	}
	docker_args = append(docker_args, containerName)
	docker_args = append(docker_args, args...)
	log.Printf("Runner docker %s", strings.Join(docker_args, " "))

	cmd := exec.Command("docker", docker_args...)

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	cmd_err := cmd.Run()
	exitStatus := 0
	if exiterr, ok := cmd_err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
			log.Printf("Exit Status: %d", exitStatus)
		}
	} else {
		log.Printf("cmd.Run: %v", cmd_err)
	}

	return exitStatus, nil
}
