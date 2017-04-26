package scheduler

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"strings"
	"time"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	ReadQueue(n int) []*tes.Task
	AssignTask(*tes.Task, *pbf.Worker)
	CheckWorkers() error
	GetWorkers(context.Context, *pbf.GetWorkersRequest) (*pbf.GetWorkersResponse, error)
	UpdateWorker(context.Context, *pbf.Worker) (*pbf.UpdateWorkerResponse, error)
}

// NewScheduler returns a new Scheduler instance.
func NewScheduler(db Database, conf config.Config) (*Scheduler, error) {
	backends := map[string]*BackendPlugin{}

	err := util.EnsureDir(conf.WorkDir)
	if err != nil {
		return nil, err
	}

	return &Scheduler{db, conf, backends}, nil
}

// Scheduler handles scheduling tasks to workers and support many backends.
type Scheduler struct {
	db       Database
	conf     config.Config
	backends map[string]*BackendPlugin
}

// AddBackend adds a backend plugin.
func (s *Scheduler) AddBackend(plugin *BackendPlugin) {
	s.backends[plugin.Name] = plugin
}

// Start starts the scheduling loop. This blocks.
//
// The scheduler will take a chunk of tasks from the queue,
// request the the configured backend schedule them, and
// act on offers made by the backend.
func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.conf.ScheduleRate)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var err error
			err = s.Schedule(ctx)
			if err != nil {
				log.Error("Schedule error", err)
				return err
			}
			err = s.Scale(ctx)
			if err != nil {
				log.Error("Scale error", err)
				return err
			}
		}
	}
}

// Schedule does a scheduling iteration. It checks the health of workers
// in the database, gets a chunk of tasks from the queue (configurable by config.ScheduleChunk),
// and calls the given scheduler. If the scheduler returns a valid offer, the
// task is assigned to the offered worker.
func (s *Scheduler) Schedule(ctx context.Context) error {
	backend, err := s.backend()
	if err != nil {
		return err
	}

	s.db.CheckWorkers()
	for _, task := range s.db.ReadQueue(s.conf.ScheduleChunk) {
		offer := backend.Schedule(task)
		if offer != nil {
			log.Info("Assigning task to worker",
				"taskID", task.Id,
				"workerID", offer.Worker.Id,
			)
			s.db.AssignTask(task, offer.Worker)
		} else {
			log.Info("No worker could be scheduled for task", "taskID", task.Id)
		}
	}
	return nil
}

// Scale implements some common logic for allowing scheduler backends
// to poll the database, looking for workers that need to be started
// and shutdown.
func (s *Scheduler) Scale(ctx context.Context) error {
	backend, err := s.backend()
	if err != nil {
		return err
	}

	b, isScaler := backend.(Scaler)
	// If the scheduler doesn't implement the Scaler interface,
	// stop here.
	if !isScaler {
		return nil
	}

	resp, err := s.db.GetWorkers(ctx, &pbf.GetWorkersRequest{})
	if err != nil {
		log.Error("Failed GetWorkers request. Recovering.", err)
		return nil
	}

	for _, w := range resp.Workers {

		if !b.ShouldStartWorker(w) {
			continue
		}

		serr := b.StartWorker(w)
		if serr != nil {
			log.Error("Error starting worker", serr)
			continue
		}

		// TODO should the Scaler instance handle this? Is it possible
		//      that Initializing is the wrong state in some cases?
		w.State = pbf.WorkerState_Initializing
		_, err := s.db.UpdateWorker(ctx, w)

		if err != nil {
			// TODO an error here would likely result in multiple workers
			//      being started unintentionally. Not sure what the best
			//      solution is. Possibly a list of failed workers.
			//
			//      If the scheduler and database API live on the same server,
			//      this *should* be very unlikely.
			log.Error("Error marking worker as initializing", err)
		}
	}
	return nil
}

// backend returns a Backend instance for the backend
// given by name in config.Scheduler.
func (s *Scheduler) backend() (Backend, error) {
	name := strings.ToLower(s.conf.Scheduler)
	plugin, ok := s.backends[name]

	if !ok {
		log.Error("Unknown scheduler backend", "name", name)
		return nil, fmt.Errorf("Unknown scheduler backend %s", name)
	}

	// Cache the scheduler instance on the plugin so that
	// we can call this backend() function repeatedly.
	if plugin.instance == nil {
		i, err := plugin.Create(s.conf)
		if err != nil {
			return nil, err
		}
		plugin.instance = i
	}
	return plugin.instance, nil
}
