package local

import (
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
	w := NewLocalWorker(conf)
	go w.Run()
	s := &scheduler{conf, w}
	return s
}

type scheduler struct {
	conf   config.Config
	worker *localWorker
}

// Schedule schedules a job and returns a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) sched.Offer {
	log.Debug("Running local scheduler")
	return sched.NewOffer(j, s.worker.Worker)
}

func NewLocalWorker(conf config.Config) *localWorker {
	id := sched.GenWorkerID()
	w := &localWorker{
		Worker: sched.Worker{
			ID: id,
			Resources: sched.Resources{
				CPU:  1,
				RAM:  5.0,
				Disk: 5.0,
			},
		},
	}
	w.Conf = conf.Worker
	w.Conf.ID = id
	w.Conf.ServerAddress = "localhost:9090"
	w.Conf.Storage = conf.Storage
	return w
}

type localWorker struct {
	sched.Worker
	Conf config.Worker
}

func (w *localWorker) Run() {
	log.Debug("Starting local worker", "storage", w.Conf.Storage)

	err := worker.Run(w.Conf)
	if err != nil {
		log.Error("Can't create worker", err)
	}
}
