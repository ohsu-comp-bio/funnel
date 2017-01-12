package condor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"tes"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	worker "tes/worker"
)

func NewScheduler(c tes.Config) sched.Scheduler {
	return &scheduler{c}
}

type scheduler struct {
	conf tes.Config
}

func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Println("Running condor scheduler")

	w := sched.Worker{
		ID: sched.GenWorkerID(),
		Resources: sched.Resources{
			// TODO
			CPU:  1,
			RAM:  1.0,
			Disk: 10.0,
		},
	}
	o := sched.NewOffer(j, w)
	go s.observe(o)
	return o
}

func (s *scheduler) observe(o sched.Offer) {
	<-o.Wait()
	if o.Accepted() {
		s.startWorker(o.Worker().ID)
	} else if o.Rejected() {
		log.Println("Condor offer was rejected")
	}
}

func (s *scheduler) startWorker(workerID string) {
	log.Println("Start condor worker")

	workerConf := worker.Config{
		ID:            workerID,
		ServerAddress: s.conf.ServerAddress,
		Timeout:       0,
		NumWorkers:    1,
		Storage:       s.conf.Storage,
	}

	confPath, cleanup := workerConf.ToYamlTempFile("worker.conf.yml")
	defer cleanup()

	workerPath := sched.DetectWorkerPath()

	condorConf := fmt.Sprintf(`
		universe = vanilla
		executable = %s
		arguments = -config worker.conf.yml
		environment = "PATH=/usr/bin"
		log = log
		error = err
		output = out
    transfer_input_files = %s
		queue
	`, workerPath, confPath)

	log.Printf("Condor submit config: \n%s", condorConf)

	cmd := exec.Command("condor_submit")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, condorConf)
	stdin.Close()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
