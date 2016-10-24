package tesTaskengineWorker

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// NewDockerEngine documentation
// TODO: documentation
func NewDockerEngine() *DockerCmd {
	return &DockerCmd{}
}

// DockerCmd documentation
// TODO: documentation
type DockerCmd struct {
}

// Run documentation
// TODO: documentation
func (dockerCmd DockerCmd) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	log.Printf("Docker Binds: %s", binds)

	dockerArgs := []string{"run", "--rm", "-i"}

	if workdir != "" {
		dockerArgs = append(dockerArgs, "-w", workdir)
	}

	for _, i := range binds {
		dockerArgs = append(dockerArgs, "-v", i)
	}
	dockerArgs = append(dockerArgs, containerName)
	dockerArgs = append(dockerArgs, args...)
	log.Printf("Runner docker %s", strings.Join(dockerArgs, " "))

	cmd := exec.Command("docker", dockerArgs...)

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	cmdErr := cmd.Run()
	exitStatus := 0
	if exiterr, ok := cmdErr.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
			log.Printf("Exit Status: %d", exitStatus)
		}
	} else {
		log.Printf("cmd.Run: %v", cmdErr)
	}

	return exitStatus, nil
}
