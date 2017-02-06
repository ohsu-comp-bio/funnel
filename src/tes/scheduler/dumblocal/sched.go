// Package dumblocal is a proof of concept scheduler showing that a scheduler could
// reuse other schedulers.
package dumblocal

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"tes"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	dumb "tes/scheduler/dumb"
)

var log = logger.New("dumbsched")

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf tes.Config) sched.Scheduler {
	return &scheduler{
		dumb.NewScheduler(conf.Schedulers.Local.NumWorkers),
		conf,
	}
}

type scheduler struct {
	dumbsched dumb.Scheduler
	conf      tes.Config
}

// Schedule schedules a job, returning a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Debug("Running dumblocal scheduler")

	o := s.dumbsched.Schedule(j)
	go s.observe(o)
	return o
}

func (s *scheduler) observe(o sched.Offer) {
	<-o.Wait()

	if o.Accepted() {
		s.dumbsched.DecrementAvailable()
		s.startWorker(o.Worker().ID)
		s.dumbsched.IncrementAvailable()

	} else if o.Rejected() {
		log.Debug("Local offer was rejected")
	}
}


func (s *scheduler) startWorker(workerID string) {
	log.Printf("Starting dumblocal worker")
	workdir := path.Join(s.conf.WorkDir, "local-scheduler", workerID)
	workdir, _ = filepath.Abs(workdir)
	os.MkdirAll(workdir, 0755)

	workerConf := s.conf.Worker
	workerConf.ID = workerID
	workerConf.ServerAddress = s.conf.ServerConfig.ServerAddress
	workerConf.Storage = s.conf.ServerConfig.Storage

	confPath := path.Join(workdir, "worker.conf.yml")
	workerConf.ToYamlFile(confPath)

	workerPath := sched.DetectWorkerPath()

	cmd := exec.Command(workerPath, "-config", confPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error("Couldn't start worker", err)
	}
}
