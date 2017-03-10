package gce

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"tes/config"
	pbr "tes/server/proto"
)

// Client is the interface the scheduler/scaler uses to interact with the GCE API.
// Mainly, the scheduler needs to be able to look up an instance template,
// or start a worker instance.
type Client interface {
	Template(project, zone, id string) (*pbr.Resources, error)
	StartWorker(project, zone, id string, conf config.Worker) error
}

func newClient(wrapper Wrapper) Client {
  return &gceClient{
    templates: map[string]*compute.InstanceTemplate{},
    wrapper: wrapper,
  }
}

// Helper for creating a wrapper before creating a client
func newClientFromConfig(conf config.Config) (Client, error) {
  w, err := newWrapper(context.Background(), conf)
  if err != nil {
    return nil, err
  }
  return newClient(w), nil
}

type gceClient struct {
	templates    map[string]*compute.InstanceTemplate
	machineTypes map[string]pbr.Resources
  wrapper      Wrapper
}

// Templates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func (s *gceClient) Template(project, zone, id string) (*pbr.Resources, error) {

	// TODO expire cache?
	if s.machineTypes == nil {
		err := s.loadMachineTypes(project, zone)
		if err != nil {
			log.Error("Couldn't load GCE machine types", err)
			return nil, err
		}
	}

	tpl, err := s.template(project, id)
	if err != nil {
		return nil, err
	}

	// Map the machine type ID string to a pbr.Resources struct
	x, ok := s.machineTypes[tpl.Properties.MachineType]

	if !ok {
		log.Error("Unknown machine type",
			"machineType", tpl.Properties.MachineType,
			"template", id)
		return nil, fmt.Errorf("Unknown machine type: %s", tpl.Properties.MachineType)
	}

	// Copy the struct so we don't modify the cached machine type data
	res := x
	// TODO is there always at least one disk? Is the first the best choice?
	//      how to know which to pick if there are multiple?
	res.Disk = float64(tpl.Properties.Disks[0].InitializeParams.DiskSizeGb)
	return &res, nil
}

// StartWorker calls out to GCE APIs to create a VM instance
// with a Funnel worker.
func (s *gceClient) StartWorker(project, zone, template string, conf config.Worker) error {

	// Get the instance template from the GCE API
	tpl, terr := s.template(project, template)
	if terr != nil {
		return terr
	}

	// TODO just put the config in the worker metadata?
	//      probably more consistent than figuring out a different config
	//      deployment for each cloud provider
	confYaml := string(conf.ToYaml())

	// Add the funnel config yaml string to the template metadata
	props := tpl.Properties
	metadata := compute.Metadata{
		Items: append(props.Metadata.Items,
			&compute.MetadataItems{
        // TODO move to conf.Metadata
				Key:   "funnel-config",
				Value: &confYaml,
			},
		),
	}

	// Prepare disk details by setting the specific zone
	for _, disk := range props.Disks {
		dt := localize(zone, "diskTypes", disk.InitializeParams.DiskType)
		disk.InitializeParams.DiskType = dt
	}

	// Create the instance on GCE
	instance := compute.Instance{
		Name:              conf.ID,
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

	op, ierr := s.wrapper.InsertInstance(project, zone, &instance)
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
		return ierr
	}

	log.Debug("GCE VM instance created", "details", op)
	return nil
}

func (s *gceClient) loadMachineTypes(project, zone string) error {
	// TODO need GCE scheduler config validation. If zone is missing, nothing works.
	// Get the machine types available to the project + zone
  resp, err := s.wrapper.ListMachineTypes(project, zone)
	if err != nil {
		log.Error("Couldn't get GCE machine list.", err)
		return err
	}
	s.machineTypes = map[string]pbr.Resources{}

	for _, m := range resp.Items {
		s.machineTypes[m.Name] = pbr.Resources{
			Cpus: uint32(m.GuestCpus),
			Ram:  float64(m.MemoryMb) / float64(1024),
		}
	}
	return nil
}

func (s *gceClient) template(project, id string) (*compute.InstanceTemplate, error) {
	// TODO expire cache?
	// Get the template from the cache, or call out to the GCE API
	tpl, exists := s.templates[id]
	if !exists {
		// Get the instance template from the GCE API
    res, err := s.wrapper.GetInstanceTemplate(project, id)
		if err != nil {
			log.Error("Couldn't get GCE template", "error", err, "template", id)
			return nil, err
		}
		tpl = res
		s.templates[id] = tpl
	}
	return tpl, nil
}

func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
