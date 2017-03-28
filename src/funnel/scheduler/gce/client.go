package gce

import (
	"context"
	"fmt"
	"funnel/config"
	pbr "funnel/server/proto"
	//"github.com/mitchellh/copystructure"
	"google.golang.org/api/compute/v1"
	"time"
)

// Client is the interface the scheduler/scaler uses to interact with the GCE API.
// Mainly, the scheduler needs to be able to look up an instance template,
// or start a worker instance.
type Client interface {
	Templates() []pbr.Worker
	StartWorker(tplName string, conf config.Worker) error
}

// Helper for creating a wrapper before creating a client
func newClientFromConfig(conf config.Config) (Client, error) {
	w, err := newWrapper(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	return &gceClient{
		wrapper:  w,
		cacheTTL: conf.Schedulers.GCE.CacheTTL,
		project:  conf.Schedulers.GCE.Project,
		zone:     conf.Schedulers.GCE.Zone,
	}, nil
}

type gceClient struct {
	// cached templates list
	templates map[string]*compute.InstanceTemplate
	// cached machine types
	machineTypes map[string]*compute.MachineType
	// GCE API wrapper
	wrapper Wrapper
	// Last time the cache was updated
	cacheTime time.Time
	// How long before expiring the cache
	cacheTTL time.Duration
	// GCE Project and Zone
	// For now at least, the client is specific to a single project + zone
	project string
	zone    string
}

// Templates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func (s *gceClient) Templates() []pbr.Worker {
	s.loadTemplates()
	workers := []pbr.Worker{}

	for id, tpl := range s.templates {

		mt, ok := s.machineTypes[tpl.Properties.MachineType]
		if !ok {
			log.Error("Couldn't find machine type. Skipping template",
				"machineType", tpl.Properties.MachineType)
			continue
		}

		disks := tpl.Properties.Disks

		res := pbr.Resources{
			Cpus: uint32(mt.GuestCpus),
			Ram:  float64(mt.MemoryMb) / float64(1024),
			// TODO is there always at least one disk? Is the first the best choice?
			//      how to know which to pick if there are multiple?
			Disk: float64(disks[0].InitializeParams.DiskSizeGb),
		}

		// Copy resources struct for available
		avail := res
		workers = append(workers, pbr.Worker{
			Resources: &res,
			Available: &avail,
			Zone:      s.zone,
			Metadata: map[string]string{
				"gce":          "yes",
				"gce-template": id,
			},
		})
	}
	return workers
}

// StartWorker calls out to GCE APIs to create a VM instance
// with a Funnel worker.
func (s *gceClient) StartWorker(tplName string, conf config.Worker) error {
	s.loadTemplates()

	// Get the instance template from the GCE API
	tpl, ok := s.templates[tplName]
	if !ok {
		return fmt.Errorf("Instance template not found: %s", tplName)
	}

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
		dt := localize(s.zone, "diskTypes", disk.InitializeParams.DiskType)
		disk.InitializeParams.DiskType = dt
	}

	// Create the instance on GCE
	instance := compute.Instance{
		Name:              conf.ID,
		CanIpForward:      props.CanIpForward,
		Description:       props.Description,
		Disks:             props.Disks,
		MachineType:       localize(s.zone, "machineTypes", props.MachineType),
		NetworkInterfaces: props.NetworkInterfaces,
		Scheduling:        props.Scheduling,
		ServiceAccounts:   props.ServiceAccounts,
		Tags:              props.Tags,
		Metadata:          &metadata,
	}

	op, ierr := s.wrapper.InsertInstance(s.project, s.zone, &instance)
	if ierr != nil {
		log.Error("Couldn't insert GCE VM instance", ierr)
		return ierr
	}

	log.Debug("GCE VM instance created", "details", op)
	return nil
}

// loadTemplates loads all the project's instance templates from the GCE API
func (s *gceClient) loadTemplates() {
	// Don't query the GCE API if we have cache results
	if s.cacheTime.IsZero() && time.Since(s.cacheTime) < s.cacheTTL {
		return
	}
	s.cacheTime = time.Now()

	// Get the machine types available
	mtresp, mterr := s.wrapper.ListMachineTypes(s.project, s.zone)
	if mterr != nil {
		log.Error("Couldn't get GCE machine list", mterr)
		return
	}

	// Get the instance template from the GCE API
	itresp, iterr := s.wrapper.ListInstanceTemplates(s.project)
	if iterr != nil {
		log.Error("Couldn't get GCE instance templates", iterr)
		return
	}

	s.machineTypes = map[string]*compute.MachineType{}
	s.templates = map[string]*compute.InstanceTemplate{}

	for _, m := range mtresp.Items {
		s.machineTypes[m.Name] = m
	}

	for _, t := range itresp.Items {
		// Only include instance templates with a "funnel" tag
		if hasTag(t) {
			s.templates[t.Name] = t
		}
	}
}

func hasTag(t *compute.InstanceTemplate) bool {
	for _, t := range t.Properties.Tags.Items {
		if t == "funnel" {
			return true
		}
	}
	return false
}

// localize helps make a resource string zone-specific
func localize(zone, resourceType, val string) string {
	return fmt.Sprintf("zones/%s/%s/%s", zone, resourceType, val)
}
