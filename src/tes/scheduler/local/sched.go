package local

import (
	"log"
	"os"
	"os/exec"
  "path"
  "path/filepath"
	"strings"
	"sync/atomic"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
	pbr "tes/server/proto"
)

// TODO Questions:
// - how to efficiently copy/slice a large resource pool?
// - how to track shutdown of workers, which free used resources in the pool?
// - how to re-evaluate the resource pool after a worker is created (autoscale)?
// - if two jobs consume parts of the same autoscale resource, how does res.Consume()
//   ensure the resource is only started once?
// - how to index resources so that scheduler can easily and efficiently match
//   a task to a resource. Don't want to loop through 1000 resources for every task
//   to find the best match. 1000 tasks and 10000 resources would be 10 million iterations.

func NewScheduler(workers int, conf []*pbr.StorageConfig) sched.Scheduler {
  // TODO HACK: get the path to the worker executable
  dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
  p := path.Join(dir, "tes-worker")
	return &scheduler{int32(workers), conf, p}
}

type scheduler struct {
	// TODO how does the pool stay updated?
	available int32
	conf      []*pbr.StorageConfig
  workerPath string
}

func (l *scheduler) Schedule(t *pbe.Task) sched.Offer {
	log.Println("Running local scheduler")

	// Make an offer if the current resource count is less than the max.
	// This is just a dumb placeholder for a future scheduler.
	//
	// A better algorithm would rank the tasks, have a concept of binpacking,
	// and be able to assign a task to a specific, best-match node.
	// This backend does none of that...yet.
	avail := atomic.LoadInt32(&l.available)
	log.Printf("Available: %d", avail)
	if avail == int32(0) {
		return sched.RejectedOffer("Pool is full")
	} else {
		w := sched.Worker{
			ID: sched.GenWorkerID(),
			Resources: sched.Resources{
				CPU:  1,
				RAM:  1.0,
				Disk: 10.0,
			},
		}
		o := sched.NewOffer(t, w)
		go l.observe(o)
		return o
	}
}

func (l *scheduler) observe(o sched.Offer) {
	<-o.Wait()
	if o.Accepted() {
		atomic.AddInt32(&l.available, -1)
		l.runWorker(o.Worker().ID)
		atomic.AddInt32(&l.available, 1)
	} else if o.Rejected() {
		log.Println("Local offer was rejected")
	}
}

func (l *scheduler) runWorker(workerID string) {
	log.Printf("Starting local worker")
	alloweddirs := make([]string, 1)
	for _, d := range l.conf {
		if d.Protocol == "fs" {
			alloweddirs = append(alloweddirs, d.Config["basedir"])
		}
	}

	cmd := exec.Command(
    l.workerPath,
		"-numworkers", "1",
		"-id", workerID,
		"-timeout", "1",
		"-alloweddirs", strings.Join(alloweddirs, ","),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("%s", err)
	}
}

//...I cannot believe I have to define these.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
