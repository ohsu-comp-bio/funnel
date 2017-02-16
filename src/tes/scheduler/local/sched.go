package local

import (
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	pbr "tes/server/proto"
	"tes/worker"
)

var log = logger.New("local")

// TODO Questions:

// - how to re-evaluate the resource pool after a worker is created (autoscale)?

// - if two jobs consume parts of the same autoscale resource, how does res.Consume()
//   ensure the resource is only started once?

// - how to index resources so that scheduler can easily and efficiently match
//   a task to a resource. Don't want to loop through 1000 resources for every task
//   to find the best match. 1000 tasks and 10000 resources would be 10 million iterations.
// - have a concept of stale resources? Could help with dead workers

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf config.Config) sched.Scheduler {
	go runLocalWorker(conf)

	t := sched.NewTracker(conf.Worker)
	go t.Run()
	return &scheduler{conf, t}
}

type scheduler struct {
	conf    config.Config
	tracker *sched.Tracker
}

func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running local scheduler")
	weights := s.conf.Schedulers.Local.Weights

	// TODO all resource tracking will break when the resources message is nil

	workers := s.tracker.Workers()
	offers := []*sched.Offer{}

	for _, w := range workers {
		// Filter out workers that don't match the job request
		// e.g. because they don't have enough resources, ports, etc.
		if !sched.Match(w, j, sched.DefaultPredicates) {
			continue
		}

		sc := sched.DefaultScores(w, j)
		sc = sc.Weighted(weights)

		offer := sched.NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching workers were found.
	if len(offers) == 0 {
		return nil
	}

	sched.SortByAverageScore(offers)
	return offers[0]
}

func runLocalWorker(conf config.Config) {
	id := sched.GenWorkerID()
	// TODO hard-coded resources
	res := &pbr.Resources{
		Cpus: 4,
		Ram:  10.0,
    Disk: 100.0,
	}

	w := config.WorkerDefaultConfig()
	w.ID = id
	w.ServerAddress = "localhost:9090"
	w.Storage = conf.Storage
	w.Resources = res
	log.Debug("Starting local worker", "storage", w.Storage)

	err := worker.Run(w)
	if err != nil {
		log.Error("Can't create worker", err)
	}
}
