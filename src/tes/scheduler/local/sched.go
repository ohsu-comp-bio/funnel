package local

import (
	"os"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	"tes/worker"
)

var log = logger.New("local")

// TODO Questions:

// - how to track job finished, which adds new available resources?

// - how to re-evaluate the resource pool after a worker is created (autoscale)?
//   - is the pool in-memory only? do workers register? do they get stored in the db?
//   - if the scheduler died and restarted, would it need to rebuild its knowledge of
//     the cluster?

// - possible to keep cluster state in memory? rebuild on failure?
//   - but how would failure capture assigned jobs?

// - if two jobs consume parts of the same autoscale resource, how does res.Consume()
//   ensure the resource is only started once?

// - how to efficiently copy/slice a large resource pool?
// - how to index resources so that scheduler can easily and efficiently match
//   a task to a resource. Don't want to loop through 1000 resources for every task
//   to find the best match. 1000 tasks and 10000 resources would be 10 million iterations.

// - have a concept of stale resources? Could help with dead workers

// - build a picture of cluster state from a log of events on every schedule?

// - reserved resources are maybe a critical concept. these are resources to which
//   a job has been assigned, but which are not yet in use (because the worker
//   hasn't picked up and started the job)
// - with reserved resources, there needs to be a way to transition those resources
//   to "active"

// - for condor, rebuilding cluster state on startup means querying condor_status

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf config.Config) sched.Scheduler {
	return &scheduler{conf, worker}
}

type scheduler struct {
	conf   config.Config
	worker sched.Worker
	avail  Resources
}

// Schedule schedules a job and returns a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running local scheduler")

	if avail == int32(0) {
		log.Debug("Worker is full")
		return nil
	}

	return sched.NewOffer(j, w)
}

// TODO start a single local worker on startup
//      even need a separate process? Can this be inprocess?
func (s *scheduler) runWorker(ctx context.Context) {
	log.Debug("Starting local worker", "storage", s.conf.Storage)

	workerConf := s.conf.Worker
	workerConf.ID = sched.GenWorkerID()
	workerConf.ServerAddress = "localhost:9090"
	workerConf.Storage = s.conf.Storage

	w, err := worker.NewWorker(conf)
	if err != nil {
		log.Error("Can't create worker", err)
	}
	ctx := context.Background()
	w.Run(ctx)

	w := sched.Worker{
		ID: sched.GenWorkerID(),
		Resources: sched.Resources{
			CPU:  1,
			RAM:  1.0,
			Disk: 10.0,
		},
	}
}
