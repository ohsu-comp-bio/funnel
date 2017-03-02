package gce

// TODO
// - how to re-evaluate the resource pool after a worker is created (autoscale)?
// - resource tracking via GCP APIs
// - provisioning
// - matching requirements to existing VMs
// - provisioning limits, e.g. don't create more than 100 VMs, or
//   maybe use N VCPUs max, across all VMs
// - act on failed machines?
// - know how to shutdown machines

// TODO outside of scheduler for SMC RNA
// - client that organizes SMCRNA tasks and submits
// - dashboard for tracking jobs
// - resource tracking via TES worker stats collection

import (
	"context"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	pbr "tes/server/proto"
)

var log = logger.New("gce")

// NewScheduler returns a new Google Cloud Engine Scheduler instance.
func NewScheduler(conf config.Config) (sched.Scheduler, error) {
	client, err := sched.NewClient(conf.Worker)
	if err != nil {
		log.Error("Can't connect scheduler client", err)
		return nil, err
	}

	gce, gerr := newGCEClient(context.TODO(), conf)
	if gerr != nil {
		log.Error("Can't connect GCE client", gerr)
		return nil, gerr
	}

	s := &scheduler{
		conf:   conf,
		client: client,
		gce:    gce,
	}

	return s, nil
}

type scheduler struct {
	conf   config.Config
	client *sched.Client
	gce    gceClientI
}

// Schedule schedules a job on a Google Cloud VM worker instance.
func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running gce scheduler")
	workers := s.getWorkers()
	weights := s.conf.Schedulers.GCE.Weights
	return sched.DefaultScheduleAlgorithm(j, workers, weights)
}

// getWorkers returns a list of all GCE workers which are not dead/gone.
// Also appends extra entries for unprovisioned workers.
func (s *scheduler) getWorkers() []*pbr.Worker {

	req := &pbr.GetWorkersRequest{}
	resp, err := s.client.GetWorkers(context.Background(), req)
	workers := []*pbr.Worker{}

	if err != nil {
		log.Error("Failed GetWorkers request. Recovering.", err)
		return workers
	}

	// Find all workers with GCE prefix in ID, that are not Dead/Gone.
	for _, w := range resp.Workers {
		if w.Gce != nil && w.State != pbr.WorkerState_Dead &&
			w.State != pbr.WorkerState_Gone {
			workers = append(workers, w)
		}
	}

	// Include unprovisioned workers.
	for _, tpl := range s.gce.Templates() {
		workers = append(workers, &pbr.Worker{
			Id:        sched.GenWorkerID(),
			Resources: tpl.Resources,
			Available: tpl.Resources,
			Zone:      s.conf.Schedulers.GCE.Zone,
			Gce: &pbr.GCEWorkerInfo{
				Template: tpl.Id,
			},
		})
	}

	return workers
}

// ShouldStartWorker tells the scaler loop which workers
// belong to this scheduler backend, basically.
func (s *scheduler) ShouldStartWorker(w *pbr.Worker) bool {
	return w.Gce != nil && w.State == pbr.WorkerState_Uninitialized
}

func (s *scheduler) StartWorker(w *pbr.Worker) error {
	return s.gce.StartWorker(w)
}
