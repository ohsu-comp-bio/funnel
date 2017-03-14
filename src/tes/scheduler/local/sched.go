package local

import (
	"golang.org/x/net/context"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	pbr "tes/server/proto"
	"tes/worker"
)

var log = logger.New("local")

// NewScheduler returns a new Scheduler instance.
func NewScheduler(conf config.Config) (sched.Scheduler, error) {
	id := sched.GenWorkerID()
	err := startWorker(id, conf)
	if err != nil {
		return nil, err
	}

	client, _ := sched.NewClient(conf.Worker)
	return &scheduler{conf, client, id}, nil
}

type scheduler struct {
	conf     config.Config
	client   sched.Client
	workerID string
}

func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running local scheduler")
	weights := s.conf.Schedulers.Local.Weights
	workers := s.getWorkers()
	return sched.DefaultScheduleAlgorithm(j, workers, weights)
}

// getWorkers gets a list of active, local workers.
//
// This is a bit redundant in the local scheduler, because there is only
// ever one worker, but it demonstrates the pattern of a scheduler backend,
// and give the scheduler a chance to get an updated worker state.
func (s *scheduler) getWorkers() []*pbr.Worker {
	workers := []*pbr.Worker{}
	resp, rerr := s.client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})

	if rerr != nil {
		log.Error("Error getting workers. Recovering.", rerr)
		return workers
	}

	for _, w := range resp.Workers {
		if w.Id != s.workerID || w.State != pbr.WorkerState_Alive {
			// Ignore workers that aren't alive
			continue
		}
		workers = append(workers, w)
	}
	return workers
}

func startWorker(id string, conf config.Config) error {
	c := conf.Worker
	c.ID = id
	c.ServerAddress = "localhost:9090"
	c.Storage = conf.Storage
	c.Resources = conf.Worker.Resources
	log.Debug("Starting local worker", "storage", c.Storage)

	w, err := worker.NewWorker(c)
	if err != nil {
		log.Error("Can't create worker", err)
		return err
	}
	go w.Run()
	return nil
}
