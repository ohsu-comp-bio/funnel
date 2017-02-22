package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

type clientI interface {
	GetWorkers(context.Context, *pbr.GetWorkersRequest, ...grpc.CallOption) (*pbr.GetWorkersResponse, error)
}

type scheduler struct {
	conf     config.Config
	client   clientI
	workerID string
}

func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running local scheduler")

	weights := s.conf.Schedulers.Local.Weights
	resp, rerr := s.client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})

	if rerr != nil {
		log.Error("Error getting workers. Recovering.", rerr)
		return nil
	}

	offers := []*sched.Offer{}

	for _, w := range resp.Workers {
		if w.Id != s.workerID || w.State != pbr.WorkerState_Alive {
			// Ignore workers that aren't alive
			continue
		}

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

	return offers[0]
}

func startWorker(id string, conf config.Config) error {
	// TODO hard-coded resources
	res := &pbr.Resources{
		Disk: 100.0,
	}

	c := conf.Worker
	c.ID = id
	c.ServerAddress = "localhost:9090"
	c.Storage = conf.Storage
	c.Resources = res
	log.Debug("Starting local worker", "storage", c.Storage)

	w, err := worker.NewWorker(c)
	if err != nil {
		log.Error("Can't create worker", err)
		return err
	}
	go w.Run()
	return nil
}
