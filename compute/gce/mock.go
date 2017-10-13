package gce

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"google.golang.org/api/compute/v1"
)

// MockBackend is a GCE backend that doesn't communicate with
// Google Cloud APIs, which is useful for testing.
type MockBackend struct {
	*Backend
	InstancesInserted []*compute.Instance
	conf              config.Config
}

// NewMockBackend returns a GCE scheduler backend that doesn't
// communicate with Google Cloud APIs,
// Useful for testing.
func NewMockBackend(conf config.Config) (*MockBackend, error) {
	// Set up a GCE scheduler backend that has a mock client
	// so that it doesn't actually communicate with GCE.
	m := &MockBackend{conf: conf}

	gceClient := &gceClient{
		wrapper: m,
		project: conf.Backends.GCE.Project,
		zone:    conf.Backends.GCE.Zone,
	}

	wpClient, err := scheduler.NewClient(conf.Scheduler)
	if err != nil {
		return nil, err
	}

	m.Backend = &Backend{
		conf:   conf,
		client: wpClient,
		gce:    gceClient,
	}
	return m, nil
}

// InsertInstance defines a mock implementation of the Wrapper interface which starts an in-memory node routine.
func (m *MockBackend) InsertInstance(project, zone string, i *compute.Instance) (*compute.Operation, error) {
	m.InstancesInserted = append(m.InstancesInserted, i)

	meta := &Metadata{}
	meta.Instance.Name = i.Name
	meta.Instance.Hostname = "localhost"

	for _, item := range i.Metadata.Items {
		if item.Key == "funnel-node-serveraddress" {
			meta.Instance.Attributes.FunnelNodeServerAddress = *item.Value
		}
	}

	meta.Instance.Zone = m.conf.Backends.GCE.Zone
	meta.Project.ProjectID = m.conf.Backends.GCE.Project
	c, cerr := WithMetadataConfig(m.conf, meta)

	if cerr != nil {
		return nil, cerr
	}

	n, err := scheduler.NewNode(c, logger.NewLogger("gce-mock-node", m.conf.Scheduler.Node.Logger))
	if err != nil {
		return nil, err
	}
	go n.Run(context.Background())
	return nil, nil
}

// ListMachineTypes defines a mock implementation of the Wrapper interface which returns a hard-coded list of machine types.
func (m *MockBackend) ListMachineTypes(proj, zone string) (*compute.MachineTypeList, error) {
	return &compute.MachineTypeList{
		Items: []*compute.MachineType{
			{
				Name:      "test-mt",
				GuestCpus: 3,
				MemoryMb:  1024,
			},
		},
	}, nil
}

// ListInstanceTemplates defines a mock implementation of the Wrapper interface which returns a hard-coded list of instance templates.
func (m *MockBackend) ListInstanceTemplates(proj string) (*compute.InstanceTemplateList, error) {
	return &compute.InstanceTemplateList{
		Items: []*compute.InstanceTemplate{
			{
				Name: "test-tpl",
				Properties: &compute.InstanceProperties{
					MachineType: "test-mt",
					Disks: []*compute.AttachedDisk{
						{
							InitializeParams: &compute.AttachedDiskInitializeParams{
								DiskSizeGb: 100,
							},
						},
					},
					Metadata: &compute.Metadata{},
					Tags: &compute.Tags{
						Items: []string{"funnel"},
					},
				},
			},
		},
	}, nil
}
