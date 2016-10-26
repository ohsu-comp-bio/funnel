package tesTaskEngineWorker

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// NewDockerEngine creates a type, DockerCmd.
func NewDockerEngine() *DockerCmd {
	return &DockerCmd{}
}

// DockerCmd has an associated method `Run`, and is an empty type.
type DockerCmd struct {
}

// Run runs a docker command.
func (dockerCmd DockerCmd) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, stdout *os.File, stderr *os.File) (int, error) {

	log.Printf("Docker Binds: %s", binds)
	// Creates docker arguments.
	dockerArgs := []string{"run", "--rm", "-i"}

	if workdir != "" {
		dockerArgs = append(dockerArgs, "-w", workdir)
	}

	for _, i := range binds {
		dockerArgs = append(dockerArgs, "-v", i)
	}

	// Append the containerName to dockerArgs.
	dockerArgs = append(dockerArgs, containerName)
	// Iterates through `args` and append it to dockerArgs.
	dockerArgs = append(dockerArgs, args...)
	log.Printf("Runner docker %s", strings.Join(dockerArgs, " "))
	// exec.Command creates a command line call, `cmd`.
	// It will look like: `run --rm -i -w [workdir] -v [bindings] [containername] [args]`
	cmd := exec.Command("docker", dockerArgs...)

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	// Runs the command line call `cmd` in the host environment.
	cmdErr := cmd.Run()
	exitStatus := 0
	if exiterr, exitOk := cmdErr.(*exec.ExitError); exitOk {
		// if exitOk is True, do the following.
		if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
			exitStatus = status.ExitStatus()
			log.Printf("Exit Status: %d", exitStatus)
		}
	} else {
		// if exitOk is False, do the following.
		log.Printf("cmd.Run: %v", cmdErr)
	}

	return exitStatus, nil
}
