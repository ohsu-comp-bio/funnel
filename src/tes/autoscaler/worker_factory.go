package autoscaler

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	pbc "tes/autoscaler/proto"
)

// TODO config
const workerCmd = "/Users/buchanae/projects/task-execution-server/bin/tes-worker"

type WorkerFactory interface {
	AddWorkers(howMany int)
}

type LocalWorkerFactory struct {
}

func (LocalWorkerFactory) AddWorkers(howMany int) {
	log.Println("Starting local worker")
	nworker := fmt.Sprintf("%d", howMany)
	cmd := exec.Command(workerCmd, "-nworkers", nworker)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Printf("%s", err)
	}
}

// CondorWorkerFactory is responsible for starting TES workers via HTCondor
// in response to an autoscaler request.
type CondorWorkerFactory struct {
	SchedAddr string
	condor    *CondorProxyClient
}

func (f CondorWorkerFactory) AddWorkers(howMany int) {
	ctx := context.Background()
	req := &pbc.StartWorkerRequest{f.SchedAddr}
	f.condor.StartWorker(ctx, req)
}
