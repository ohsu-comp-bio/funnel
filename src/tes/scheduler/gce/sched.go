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

	s := &scheduler{
		conf:   conf,
		client: client,
		// Available worker types are described using GCE instance templates.
		templates: []*pbr.Worker{},
	}

	// Factory watches for new workers in the Funnel database
	// and provisions new GCE VMs as necessary.
	//f, _ := newFactory(conf)
	//go f.Run()

	return s, nil
}

type scheduler struct {
	conf      config.Config
	client    *sched.Client
	templates []*pbr.Worker
}

// Schedule schedules a job on a Google Cloud VM worker instance.
func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running gce scheduler")

	weights := s.conf.Schedulers.GCE.Weights
	offers := []*sched.Offer{}

	for _, w := range s.getWorkers() {
		// Filter out workers that don't match the job request.
		// Checks CPU, RAM, disk space, ports, etc.
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
		if w.Gce != nil && w.State != pbr.WorkerState_Dead && w.State != pbr.WorkerState_Gone {
			workers = append(workers, w)
		}
	}

	// Include unprovisioned workers.
	for _, tpl := range s.templates {
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

// getTemplates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func getTemplates(conf config.Config) []*pbr.Worker {
	templates := []*pbr.Worker{}
	project := conf.Schedulers.GCE.Project

	// TODO mockable for testing
	svc, serr := service(context.Background(), conf)
	if serr != nil {
		return templates
	}

	machineTypes := map[string]pbr.Resources{}

	// Get the machine types available to the project + zone
	resp, err := svc.MachineTypes.List(project, conf.Schedulers.GCE.Zone).Do()
	if err != nil {
		log.Error("Couldn't get GCE machine list.")
		// TODO return error?
		return templates
	}

	for _, m := range resp.Items {
		machineTypes[m.Name] = pbr.Resources{
			Cpus: uint32(m.GuestCpus),
			Ram:  float64(m.MemoryMb) / float64(1024),
		}
	}

	for _, t := range conf.Schedulers.GCE.Templates {
		// Get the instance template from the GCE API
		tpl, err := svc.InstanceTemplates.Get(project, t).Do()
		if err != nil {
			log.Error("Couldn't get GCE template. Skipping.", "error", err, "template", t)
			continue
		}
		// Map the machine type ID string to a pbr.Resources struct
		res := machineTypes[tpl.Properties.MachineType]
		// TODO is there always at least one disk? Is the first the best choice?
		//      how to know which to pick if there are multiple?
		res.Disk = float64(tpl.Properties.Disks[0].InitializeParams.DiskSizeGb)
		templates = append(templates, &pbr.Worker{
			Resources: &res,
		})
	}
	return templates
}
