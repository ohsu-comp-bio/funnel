package condor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(s.conf.WorkDir, "condor-scheduler", workerID)
	workdir, _ = filepath.Abs(workdir)
	os.MkdirAll(workdir, 0755)

	workerConf := worker.Config{
		ID:            workerID,
		ServerAddress: s.conf.ServerAddress,
		Timeout:       0,
		NumWorkers:    1,
		Storage:       s.conf.Storage,
		WorkDir:       workdir,
	}
	confPath := path.Join(workdir, "worker.conf.yml")
	workerConf.ToYamlFile(confPath)

	workerPath := sched.DetectWorkerPath()

	condorConf := fmt.Sprintf(`
		universe = vanilla
		executable = %s
		arguments = -config worker.conf.yml
		environment = "PATH=/usr/bin"
		log = condor-event-log
		error = tes-worker-stderr
		output = tes-worker-stdout
    transfer_input_files = %s
    initial_dir = %s
		queue
	`, workerPath, confPath, workdir)

	log.Printf("Condor submit config: \n%s", condorConf)

	cmd := exec.Command("condor_submit")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, condorConf)
	stdin.Close()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
