package local

import (
	"os"
	"os/exec"
	"sync/atomic"
	"tes"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	worker "tes/worker"
)

var log = logger.New("local-sched")

// TODO Questions:
// - how to efficiently copy/slice a large resource pool?
// - how to track shutdown of workers, which free used resources in the pool?
// - how to re-evaluate the resource pool after a worker is created (autoscale)?
// - if two jobs consume parts of the same autoscale resource, how does res.Consume()
//   ensure the resource is only started once?
// - how to index resources so that scheduler can easily and efficiently match
//   a task to a resource. Don't want to loop through 1000 resources for every task
//   to find the best match. 1000 tasks and 10000 resources would be 10 million iterations.

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf tes.Config) sched.Scheduler {
	return &scheduler{
		conf,
		int32(conf.Schedulers.Local.NumWorkers),
	}
}

type scheduler struct {
	// TODO how does the pool stay updated?
	conf      tes.Config
	available int32
}

// Schedule schedules a job and returns a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Debug("Running local scheduler")

	// Make an offer if the current resource count is less than the max.
	// This is just a dumb placeholder for a future scheduler.
	//
	// A better algorithm would rank the jobs, have a concept of binpacking,
	// and be able to assign a job to a specific, best-match node.
	// This backend does none of that...yet.
	avail := atomic.LoadInt32(&s.available)
	log.Debug("Available", "slots", avail)
	if avail == int32(0) {
		return sched.RejectedOffer("Pool is full")
	}

	w := sched.Worker{
		ID: sched.GenWorkerID(),
		Resources: sched.Resources{
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
		atomic.AddInt32(&s.available, -1)
		s.runWorker(o.Worker().ID)
		atomic.AddInt32(&s.available, 1)
	} else if o.Rejected() {
		log.Debug("Local offer was rejected")
	}
}

func (s *scheduler) runWorker(workerID string) {
	log.Debug("Starting local worker", "storage", s.conf.Storage)

	workerConf := worker.Config{
		ID:            workerID,
		ServerAddress: s.conf.ServerAddress,
		Timeout:       1,
		NumWorkers:    1,
		Storage:       s.conf.Storage,
		WorkDir:       s.conf.WorkDir,
	}

	confPath, cleanup := workerConf.ToYamlTempFile("worker.conf.yml")
	defer cleanup()

	workerPath := sched.DetectWorkerPath()

	cmd := exec.Command(
		workerPath,
		"-config", confPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error("Couldn't start local worker", err)
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
