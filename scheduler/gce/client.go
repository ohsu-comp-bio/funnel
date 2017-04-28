package gce

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	//"github.com/mitchellh/copystructure"
	"google.golang.org/api/compute/v1"
	"time"
)

// Client is the interface the scheduler/scaler uses to interact with the GCE API.
// Mainly, the scheduler needs to be able to look up an instance template,
// or start a worker instance.
type Client interface {
	Templates() []pbf.Worker
	StartWorker(tplName, serverAddress, workerID string) error
}

// Helper for creating a wrapper before creating a client
func newClientFromConfig(conf config.Config) (Client, error) {
	w, err := newWrapper(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	return &gceClient{
		wrapper:  w,
		cacheTTL: conf.Backends.GCE.CacheTTL,
		project:  conf.Backends.GCE.Project,
		zone:     conf.Backends.GCE.Zone,
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
func (s *gceClient) Templates() []pbf.Worker {
	s.loadTemplates()
	workers := []pbf.Worker{}

	for id, tpl := range s.templates {

		mt, ok := s.machineTypes[tpl.Properties.MachineType]
		if !ok {
			log.Error("Couldn't find machine type. Skipping template",
				"machineType", tpl.Properties.MachineType)
			continue
		}

		disks := tpl.Properties.Disks

		res := pbf.Resources{
			Cpus:  uint32(mt.GuestCpus),
			RamGb: float64(mt.MemoryMb) / float64(1024),
			// TODO is there always at least one disk? Is the first the best choice?
			//      how to know which to pick if there are multiple?
			DiskGb: float64(disks[0].InitializeParams.DiskSizeGb),
		}

		// Copy resources struct for available
		avail := res
		workers = append(workers, pbf.Worker{
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
func (s *gceClient) StartWorker(tplName, serverAddress, workerID string) error {
	s.loadTemplates()

	// Get the instance template from the GCE API
	tpl, ok := s.templates[tplName]
	if !ok {
		return fmt.Errorf("Instance template not found: %s", tplName)
	}

	// Add GCE instance metadata
	props := tpl.Properties
	metadata := compute.Metadata{
		Items: append(props.Metadata.Items,
			&compute.MetadataItems{
				Key:   "funnel-worker-serveraddress",
				Value: &serverAddress,
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
		Name:              workerID,
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
		log.Error("Couldn't get GCE machine list",
			"error", mterr,
			"project", s.project,
			"zone", s.zone)
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
