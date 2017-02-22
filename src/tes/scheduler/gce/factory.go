package gce

/*
import (
	"fmt"
	"google.golang.org/api/compute/v1"
	"tes/config"
)

type WorkerFactory struct {
  client *sched.Client
}

func NewFactory(conf config.Config) (sched.Factory, error) {
  // Get Google API client
	svc, serr := service(context.Background(), conf)
	if serr != nil {
		log.Error("Couldn't create GCE service client", serr)
		return nil, serr
	}
  f := &factory{conf, svc}
  return f, nil
}

type factory struct {
	conf    config.Config
  client  *sched.Client
}


type watcher struct {
  client *sched.Client
  conf   config.Config
}

func WatchWorkers(conf config.Config, query *pbr.GetWorkersRequest) {
	client, _ := sched.NewClient(conf.Worker)
  workers := []*pbr.Worker{}
	ticker := time.NewTicker(conf.WatchPollRate)

	for {
		<-ticker.C

		resp, err := t.client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})
		if err != nil {
			log.Error("Failed GetWorkers request. Recovering.", err)
			continue
		}
	}
}

func (t *tracker) SetWorkerInitializing(id string) {
  _, err := t.client.SetWorkerState(context.Background(), &pbr.SetWorkerStateRequest{
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
}

func (s *scheduler) startWorker(workerID string) {
	project := s.conf.Schedulers.GCE.Project
  // TODO can zone be template specific?
	zone := s.conf.Schedulers.GCE.Zone

  // Get the instance template from the GCE API
	tpl, terr := s.svc.InstanceTemplates.Get(project, s.conf.Schedulers.GCE.Template).Do()
	if terr != nil {
		log.Error("Couldn't retrieve GCE instance template",
			"error", terr,
			"template", s.conf.Schedulers.GCE.Template)
		return
	}

  // TODO
  confYaml := ""

  // Add the funnel config yaml string to the template metadata
	props := tpl.Properties
	metadata := compute.Metadata{
		Items: append(props.Metadata.Items,
			&compute.MetadataItems{
				Key:   "funnel-config",
				Value: &confYaml,
			},
		),
	}

	for _, disk := range props.Disks {
		disk.InitializeParams.DiskType = localize(zone, "diskTypes", disk.InitializeParams.DiskType)
	}

	instance := compute.Instance{
		Name:              workerID,
		CanIpForward:      props.CanIpForward,
		Description:       props.Description,
		Disks:             props.Disks,
		MachineType:       localize(zone, "machineTypes", props.MachineType),
		NetworkInterfaces: props.NetworkInterfaces,
		Scheduling:        props.Scheduling,
		ServiceAccounts:   props.ServiceAccounts,
		Tags:              props.Tags,
		Metadata:          &metadata,
	}

	op, ierr := s.svc.Instances.Insert(project, zone, &instance).Do()
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
	}
	log.Debug("VM instance", op)
}

func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
*/
