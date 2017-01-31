package condor

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"tes"
  "tes/logger"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	worker "tes/worker"
	"text/template"
)

var log logger.Logger = logger.New("condor-sched")

// NewScheduler returns a new HTCondor Scheduler instance.
func NewScheduler(c tes.Config) sched.Scheduler {
	return &scheduler{c}
}

type scheduler struct {
	conf tes.Config
}

// Schedule schedules a job on the HTCondor queue and returns a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Debug("Running condor scheduler")

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
		log.Debug("Condor offer was rejected")
	}
}

func (s *scheduler) startWorker(workerID string) {
	log.Debug("Start condor worker")
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
		WorkDir:       "",
	}
	confPath := path.Join(workdir, "worker.conf.yml")
	workerConf.ToYamlFile(confPath)

	workerPath := sched.DetectWorkerPath()

	submitPath := path.Join(workdir, "condor.submit")
	f, _ := os.Create(submitPath)

	submitTpl, _ := template.New("condor.submit").Parse(`
		universe    = vanilla
		executable  = {{.Executable}}
		arguments   = -config worker.conf.yml
		environment = "PATH=/usr/bin"
		log         = {{.WorkDir}}/condor-event-log
		error       = {{.WorkDir}}/tes-worker-stderr
		output      = {{.WorkDir}}/tes-worker-stdout
    input       = {{.Config}}
    should_transfer_files   = YES
    when_to_transfer_output = ON_EXIT
		queue
	`)
	submitTpl.Execute(f, map[string]string{
		"Executable": workerPath,
		"WorkDir":    workdir,
		"Config":     confPath,
	})
	f.Close()

	cmd := exec.Command("condor_submit")
	stdin, _ := os.Open(submitPath)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
