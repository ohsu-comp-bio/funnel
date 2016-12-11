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

func (dockerCmd DockerCmd) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, stdin *os.File, stdout *os.File, stderr *os.File) (int, error) {

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
	log.Printf("Running command: docker %s", strings.Join(dockerArgs, " "))
	// exec.Command creates a command line call, `cmd`.
	// It will look like: `run --rm -i -w [workdir] -v [bindings] [containername] [args]`
	cmd := exec.Command("docker", dockerArgs...)

	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	// Runs the command line call `cmd` in the host environment.
	err := cmd.Run()
	exitStatus := getExitStatus(err)
	log.Printf("Exit Status: %d", exitStatus)

	return exitStatus, err
}

// getExitStatus gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitStatus(err error) int {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return status.ExitStatus()
			}
		} else {
			log.Printf("Could not determine exit code. Using default -999")
			return -999
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
}
