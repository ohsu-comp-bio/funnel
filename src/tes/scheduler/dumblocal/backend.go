// This is a proof of concept that a scheduler could be composed of another scheduler,
// and perhaps other composable pieces, such as worker factories and resource pools.
package dumblocal

import (
  "log"
	"os"
	"os/exec"
  "tes/scheduler/dumb"
  sched "tes/scheduler"
  pbe "tes/ga4gh"
)

// TODO config
const workerCmd = "/Users/buchanae/projects/task-execution-server/bin/tes-worker"

func NewScheduler(workers int) sched.Scheduler {
  return &scheduler{dumb.NewScheduler(workers)}
}

type scheduler struct {
  dumbsched Scheduler
}

func (s *scheduler) Schedule(t *pbe.Task) sched.Offer {
  log.Println("Running dumblocal scheduler")

  o := s.dumbsched.Schedule(t)
  go s.observe(o)
  return o
}

func (s *scheduler) observe(o sched.Offer) {
  <-o.Wait()

  if o.Accepted() {
    s.dumbsched.DecrementAvailable()
    runWorker(w)
    s.dumbsched.IncrementAvailable()

  } else if o.Rejected() {
    log.Println("Local offer was rejected")
  }
}

func runWorker() {
  log.Printf("Starting dumblocal worker")
  cmd := exec.Command(workerCmd, "-numworkers", "1", "-id", w.ID, "-timeout", "0")
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  err := cmd.Run()
  if err != nil {
    log.Printf("%s", err)
  }
}
