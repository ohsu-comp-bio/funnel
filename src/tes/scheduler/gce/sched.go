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
	"strings"
	"sync"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	pbr "tes/server/proto"
	"time"
)

var log = logger.New("gce")

const prefix = "gce-"

func genWorkerID() string {
	return prefix + sched.GenWorkerID()
}

// NewScheduler returns a new Google Cloud Engine Scheduler instance.
func NewScheduler(conf config.Config) sched.Scheduler {
	workers := []*pbr.Worker{}
	s := &scheduler{
		conf:    conf,
		workers: workers,
	}
	go s.track()
	return s
}

type scheduler struct {
	conf    config.Config
	workers []*pbr.Worker
	mtx     sync.Mutex
}

// track helps the scheduler know when a job has been assigned to a gce worker,
// so that the worker can be provisioned.
//
// This polls the server, looking for jobs which are assigned to a worker with
// a "gce-worker-" ID prefix. When such a worker is found, if it has an
// assigned (inactive) job, a worker is provisioned.
//
// TODO this code is completely duplicated with the condor scheduler backend
func (s *scheduler) track() {
	client, _ := sched.NewClient(s.conf.Worker)
	defer client.Close()

	ticker := time.NewTicker(s.conf.Worker.TrackerRate)

	for {
		<-ticker.C
		log.Debug("TICK")

		workers := []*pbr.Worker{}

		// TODO allow GetWorkers() to include query for prefix and state
		resp, err := client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})
		if err != nil {
			log.Error("Failed GetWorkers request. Recovering.", err)
			continue
		}

		for _, w := range resp.Workers {
			log.Debug("Checking worker", "id", w.Id, "state", w.State)

			if w.State == pbr.WorkerState_Dead || w.State == pbr.WorkerState_Gone {
				continue
			}

			if strings.HasPrefix(w.Id, prefix) &&
				w.State == pbr.WorkerState_Unknown &&
				len(w.Assigned) > 0 {

				log.Debug("Starting worker", "id", w.Id)
				s.startWorker(w.Id)

				_, err := client.SetWorkerState(context.Background(), &pbr.SetWorkerStateRequest{
					Id:    w.Id,
					State: pbr.WorkerState_Initializing,
				})

				if err != nil {
					// TODO how to handle error? On the next loop, we'll accidentally start
					//      the worker again, because the state will be Unknown still.
					//
					//      keep a local list of failed workers?
					log.Error("Can't set worker state to initialzing.", err)
				}

				w.State = pbr.WorkerState_Initializing
				workers = append(workers, w)
			}
		}

		// Fill in available but un-provisioned workers
		for i := len(workers); i < s.conf.Schedulers.GCE.MaxWorkers; i++ {
			res := &pbr.Resources{
				// TODO pull from template
				Cpus: 8,
				Ram:  8.0,
			}
			workers = append(workers, &pbr.Worker{
				Id:        genWorkerID(),
				Resources: res,
				Available: res,
				Zone:      s.conf.Schedulers.GCE.Zone,
			})
		}

		// log.Debug("Updated workers in track", workers)
		s.mtx.Lock()
		s.workers = workers
		s.mtx.Unlock()
	}
}

// Schedule schedules a job on a Google Cloud VM worker instance.
func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running gce scheduler")
	weights := s.conf.Schedulers.GCE.Weights

	s.mtx.Lock()
	workers := make([]*pbr.Worker, len(s.workers))
	copy(workers, s.workers)
	s.mtx.Unlock()

	log.Debug("GCE sched workers", len(workers))
	offers := []*sched.Offer{}

	for _, w := range workers {
		// Filter out workers that don't match the job request
		// e.g. because they don't have enough resources, ports, etc.
		if !sched.Match(w, j, sched.DefaultPredicates) {
			// TODO allow easily debugging why a worker doesn't fit
			log.Debug("Worker doesn't fit")
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
