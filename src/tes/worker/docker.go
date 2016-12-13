package tesTaskEngineWorker

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type DockerCmd struct {
	ImageName       string
	Cmd             []string
	Binds           []string
	Workdir         string
	RemoveContainer bool
	Stdin           *os.File
	Stdout          *os.File
	Stderr          *os.File
}

func (dcmd DockerCmd) Run() error {
	log.Printf("Docker Binds: %s", dcmd.Binds)

	// TODO why are we using '-i'?
	args := []string{"run", "-i"}

	if dcmd.RemoveContainer {
		args = append(args, "--rm")
	}

	if dcmd.Workdir != "" {
		args = append(args, "-w", dcmd.Workdir)
	}

	for _, bind := range dcmd.Binds {
		args = append(args, "-v", bind)
	}

	args = append(args, dcmd.ImageName)
	args = append(args, dcmd.Cmd...)

	log.Printf("Running command: docker %s", strings.Join(args, " "))
	// Roughly: `docker run --rm -i -w [workdir] -v [bindings] [imageName] [cmd]`
	cmd := exec.Command("docker", args...)

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
