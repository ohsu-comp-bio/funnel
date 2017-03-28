package local

import (
	"funnel/config"
	pbe "funnel/ga4gh"
	"funnel/logger"
	sched "funnel/scheduler"
	pbr "funnel/server/proto"
	"funnel/worker"
	"golang.org/x/net/context"
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
	c.Timeout = -1

	log.Debug("Starting local worker", "storage", c.Storage)

	w, err := worker.NewWorker(c)
	if err != nil {
		log.Error("Can't create worker", err)
		return err
	}
	go w.Run()
	return nil
}
