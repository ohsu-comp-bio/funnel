package gce

import (
	"errors"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	gce_mocks "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/compute/v1"
	"testing"
)

func init() {
	logger.Configure(logger.DebugConfig())
}

// Tests what happens when the InsertInstance() call fails the first couple times.
func TestInsertTempError(t *testing.T) {

	conf := config.DefaultConfig()
	conf.Backends.GCE.Project = "test-proj"
	conf.Backends.GCE.Zone = "test-zone"
	conf.Worker.ID = "test-worker"
	wpr := new(gce_mocks.Wrapper)
	wpr.SetupMockInstanceTemplates()
	wpr.SetupMockMachineTypes()
	client := &gceClient{
		wrapper: wpr,
		project: "test-proj",
		zone:    "test-zone",
	}

	// Set InsertInstance to return an error
	wpr.On("InsertInstance", "test-proj", "test-zone", mock.Anything).Return(nil, errors.New("TEST"))
	// Try to start the worker a few times
	// Do this a few times to exacerbate any errors.
	// e.g. a previous bug would build up a longer config string after every failure
	//      because cached data was being incorrectly shared.
	client.StartWorker("test-tpl", conf.Worker.ServerAddress, conf.Worker.ID)
	client.StartWorker("test-tpl", conf.Worker.ServerAddress, conf.Worker.ID)
	client.StartWorker("test-tpl", conf.Worker.ServerAddress, conf.Worker.ID)
	wpr.AssertExpectations(t)
	// Clear the previous expected calls
	wpr.ExpectedCalls = nil
	wpr.SetupMockInstanceTemplates()
	wpr.SetupMockMachineTypes()

	// Now set InsertInstance to success
	addr := conf.RPCAddress()
	expected := &compute.Instance{
		// TODO test that these fields get passed through from the template correctly.
		//      i.e. mock a more complex template
		CanIpForward:      false,
		CpuPlatform:       "",
		CreationTimestamp: "",
		Description:       "",
		Disks: []*compute.AttachedDisk{
			{
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskSizeGb: 100,
					DiskType:   "zones/test-zone/diskTypes/", // TODO??? this must be wrong
				},
			},
		},
		Name:        "test-worker",
		MachineType: "zones/test-zone/machineTypes/test-mt",
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "funnel-worker-serveraddress",
					Value: &addr,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"funnel"},
		},
	}
	wpr.On("InsertInstance", "test-proj", "test-zone", expected).Return(nil, nil)

	client.StartWorker("test-tpl", conf.Worker.ServerAddress, conf.Worker.ID)
	wpr.AssertExpectations(t)
}
