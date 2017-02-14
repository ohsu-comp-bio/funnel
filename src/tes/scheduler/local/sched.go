package local

import (
	"sort"
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

	workers := []*sched.Worker{}
	workers = append(workers, w.Worker)

	t := &tracker{workers}
	return &scheduler{conf, t}
}

type scheduler struct {
	conf    config.Config
	tracker *tracker
}

func (s *scheduler) Schedule(j *pbe.Job) *sched.Worker {
	log.Debug("Running local scheduler")

	// TODO all resource tracking will break when the resources message is nil

	workers := s.tracker.Get()
	fit := sched.Fit(j, workers)

	if len(fit) == 0 {
		return nil
	}

	for _, x := range fit {
		x.CalcScores(j)
	}

	// TODO explore how a scheduler could have custom scores
	//      probably just a custom worker with sched.Worker embedded
	sort.Sort(sched.ByScore(fit))
	w := fit[0]

	return w
}

// tracker helps poll the database for updated worker information.
// TODO consider making an interface for this, which a condor track would implement
type tracker struct {
	workers []*sched.Worker
}

func (t *tracker) Get() []*sched.Worker {
	return t.workers
}

func NewLocalWorker(conf config.Config) *localWorker {
	id := sched.GenWorkerID()
	w := &localWorker{
		&sched.Worker{
			ID: id,
			Resources: sched.Resources{
				CPUs: 2,
				RAM:  10.0,
				//Disk: 10.0,
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
	*sched.Worker
}

func (w *localWorker) Run() {
	log.Debug("Starting local worker", "storage", w.Conf.Storage)

	err := worker.Run(w.Conf)
	if err != nil {
		log.Error("Can't create worker", err)
	}
}
