// This is a proof of concept that a scheduler could be composed of another scheduler,
// and perhaps other composable pieces, such as worker factories and resource pools.
package dumblocal

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	dumb "tes/scheduler/dumb"
)

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

func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Println("Running dumblocal scheduler")

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
		log.Println("Local offer was rejected")
	}
}

func runWorker(workerID string, workerPath string) {
	log.Printf("Starting dumblocal worker")
	cmd := exec.Command(workerPath, "-numworkers", "1", "-id", workerID, "-timeout", "0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("%s", err)
	}
}
