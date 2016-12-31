package local

import (
	"log"
	"os"
	"os/exec"
)

// TODO config
const workerCmd = "/Users/buchanae/projects/task-execution-server/bin/tes-worker"

// Factory is responsible for starting workers by submitting jobs to HTCondor.
type factory struct {}

func (factory) StartWorker(id string) {
	log.Println("Starting local worker")
	cmd := exec.Command(workerCmd, "-numworkers", "1", "-id", id)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Printf("%s", err)
	}
}
