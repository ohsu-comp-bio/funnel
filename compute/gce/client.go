package gce

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"google.golang.org/api/compute/v1"
	"time"
)

// Client is the interface the scheduler/scaler uses to interact with the GCE API.
// Mainly, the scheduler needs to be able to look up an instance template,
// or start a node instance.
type Client interface {
	Templates() []pbs.Node
	StartNode(tplName, serverAddress, nodeID string) error
}

// Helper for creating a wrapper before creating a client
func newClientFromConfig(conf config.Config) (*gceClient, error) {
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
	log     *logger.Logger
}

// Templates queries the GCE API to get details about GCE instance templates.
// If the API client fails to connect, this returns an empty list.
func (s *gceClient) Templates() []pbs.Node {
	loaderr := s.loadTemplates()
	if loaderr != nil {
		s.log.Error("error loading GCE templates", loaderr)
	}
	nodes := []pbs.Node{}

	for id, tpl := range s.templates {

		mt, ok := s.machineTypes[tpl.Properties.MachineType]
		if !ok {
			s.log.Error("error finding GCE template machine type",
				"name", tpl.Properties.MachineType)
			continue
		}

		disks := tpl.Properties.Disks

		res := pbs.Resources{
			Cpus:  uint32(mt.GuestCpus),
			RamGb: float64(mt.MemoryMb) / float64(1024),
			// TODO is there always at least one disk? Is the first the best choice?
			//      how to know which to pick if there are multiple?
			DiskGb: float64(disks[0].InitializeParams.DiskSizeGb),
		}

		// Copy resources struct for available
		avail := res
		nodes = append(nodes, pbs.Node{
			Resources: &res,
			Available: &avail,
			Zone:      s.zone,
			Metadata: map[string]string{
				"gce":          "yes",
				"gce-template": id,
			},
		})
	}
	return nodes
}

// StartNode calls out to GCE APIs to create a VM instance
// with a Funnel node.
func (s *gceClient) StartNode(tplName, serverAddress, nodeID string) error {
	loaderr := s.loadTemplates()
	if loaderr != nil {
		s.log.Error("error loading GCE templates", loaderr)
	}

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
				Key:   "funnel-node-serveraddress",
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
		Name:              nodeID,
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

	_, ierr := s.wrapper.InsertInstance(s.project, s.zone, &instance)
	if ierr != nil {
		return fmt.Errorf("Couldn't insert GCE VM instance: %s", ierr)
	}

	return nil
}

// loadTemplates loads all the project's instance templates from the GCE API
func (s *gceClient) loadTemplates() error {
	// Don't query the GCE API if we have cache results
	if s.cacheTime.IsZero() && time.Since(s.cacheTime) < s.cacheTTL {
		return nil
	}
	s.cacheTime = time.Now()

	// Get the machine types available
	mtresp, mterr := s.wrapper.ListMachineTypes(s.project, s.zone)
	if mterr != nil {
		return mterr
	}

	// Get the instance template from the GCE API
	itresp, iterr := s.wrapper.ListInstanceTemplates(s.project)
	if iterr != nil {
		return iterr
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
	return nil
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
