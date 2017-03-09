package gce

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
	"tes/config"
	pbr "tes/server/proto"
)

// Client is the interface used by the scheduler, which is expected to
// communicate with Google Cloud Platform APIs. Mostly, this exists for
// testing. It's easier to mock this small, specific interface.
type Client interface {
	Template(project, id string) (*pbr.Resources, error)
	StartWorker(project, zone, id string, conf config.Worker) error
}

// service creates a Google Compute service client
func newClient(ctx context.Context, conf config.Config) (Client, error) {
	var client *http.Client
	if conf.Schedulers.GCE.AccountFile != "" {
		// Pull the client configuration (e.g. auth) from a given account file.
		// This is likely downloaded from Google Cloud manually via IAM & Admin > Service accounts.
		bytes, rerr := ioutil.ReadFile(conf.Schedulers.GCE.AccountFile)
		if rerr != nil {
			return nil, rerr
		}

		config, tserr := google.JWTConfigFromJSON(bytes, compute.ComputeScope)
		if tserr != nil {
			return nil, tserr
		}
		client = config.Client(ctx)
	} else {
		// Pull the information (auth and other config) from the environment,
		// which is useful when this code is running in a Google Compute instance.
		client, _ = google.DefaultClient(ctx, compute.ComputeScope)
		// TODO catch error
	}

	svc, cerr := compute.New(client)
	if cerr != nil {
		return nil, cerr
	}

	return &gceClient{
		conf: conf,
		svc:  svc,
	}, nil
}

type gceClient struct {
	templates    map[string]*compute.InstanceTemplate
	machineTypes map[string]pbr.Resources
	conf         config.Config
	svc          *compute.Service
}

// Templates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func (s *gceClient) Template(project, id string) (*pbr.Resources, error) {

	// TODO expire cache?
	if s.machineTypes == nil {
		err := s.loadMachineTypes(project)
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

func (s *gceClient) template(project, id string) (*compute.InstanceTemplate, error) {
	// TODO expire cache?
	// Get the template from the cache, or call out to the GCE API
	tpl, exists := s.templates[id]
	if !exists {
		// Get the instance template from the GCE API
		res, err := s.svc.InstanceTemplates.Get(project, id).Do()
		if err != nil {
			log.Error("Couldn't get GCE template", "error", err, "template", id)
			return nil, err
		}
		tpl = res
		s.templates[id] = tpl
	}
	return tpl, nil
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

	op, ierr := s.svc.Instances.Insert(project, zone, &instance).Do()
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
		return ierr
	}

	log.Debug("GCE VM instance created", "details", op)
	return nil
}

func (s *gceClient) loadMachineTypes(project string) error {
	// TODO need GCE scheduler config validation. If zone is missing, nothing works.
	// Get the machine types available to the project + zone
	resp, err := s.svc.MachineTypes.List(project, s.conf.Schedulers.GCE.Zone).Do()
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

func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
