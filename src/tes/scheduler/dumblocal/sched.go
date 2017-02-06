// Package dumblocal is a proof of concept scheduler showing that a scheduler could
// reuse other schedulers.
package dumblocal

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	dumb "tes/scheduler/dumb"
)

var log = logger.New("dumbsched")

// NewScheduler returns a new Scheduler instance.
func NewScheduler(workers int) sched.Scheduler {
	// TODO HACK: get the path to the worker executable
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	p := path.Join(dir, "tes-worker")
	return &scheduler{dumb.NewScheduler(workers), p}
}

type scheduler struct {
	dumbsched  dumb.Scheduler
	workerPath string
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
		runWorker(o.Worker().ID, s.workerPath)
		s.dumbsched.IncrementAvailable()

	} else if o.Rejected() {
		log.Debug("Local offer was rejected")
	}
}

func runWorker(workerID string, workerPath string) {
	log.Debug("Starting dumblocal worker")
	cmd := exec.Command(workerPath, "-numworkers", "1", "-id", workerID, "-timeout", "0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error("Couldn't start worker", err)
	}
}
