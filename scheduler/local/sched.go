package local

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/worker"
	"golang.org/x/net/context"
)

var log = logger.New("local")

// Plugin provides the local scheduler backend plugin
var Plugin = &scheduler.BackendPlugin{
	Name:   "local",
	Create: NewBackend,
}

// NewBackend returns a new Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	id := scheduler.GenWorkerID("local")
	err := startWorker(id, conf)
	if err != nil {
		return nil, err
	}

	client, _ := scheduler.NewClient(conf.Worker)
	return scheduler.Backend(&Backend{conf, client, id}), nil
}

// Backend represents the local backend.
type Backend struct {
	conf     config.Config
	client   scheduler.Client
	workerID string
}

// Schedule schedules a task.
func (s *Backend) Schedule(j *tes.Task) *scheduler.Offer {
	log.Debug("Running local scheduler backend")
	weights := map[string]float32{}
	workers := s.getWorkers()
	return scheduler.DefaultScheduleAlgorithm(j, workers, weights)
}

// getWorkers gets a list of active, local workers.
//
// This is a bit redundant in the local scheduler, because there is only
// ever one worker, but it demonstrates the pattern of a scheduler backend,
// and give the scheduler a chance to get an updated worker state.
func (s *Backend) getWorkers() []*pbf.Worker {
	workers := []*pbf.Worker{}
	resp, rerr := s.client.GetWorkers(context.Background(), &pbf.GetWorkersRequest{})

	if rerr != nil {
		log.Error("Error getting workers. Recovering.", rerr)
		return workers
	}

	for _, w := range resp.Workers {
		if w.Id != s.workerID || w.State != pbf.WorkerState_Alive {
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
