package gce

import (
	"context"
	"errors"
	"funnel/logger"
	"funnel/scheduler"
	gce_mocks "funnel/scheduler/gce/mocks"
	server_mocks "funnel/server/mocks"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/compute/v1"
	"testing"
)

func init() {
	logger.ForceColors()
}

// Test the scheduler while mocking out a lower level than what's in sched_test.go
// This mocks out the Wrapper interface, which allows better testing the logic
// in client.go and allows a more end-to-end test.
func TestWrapper(t *testing.T) {

	ctx := context.Background()
	// Mock config
	conf := basicConf()

	// Mock the GCE API wrapper
	wpr := new(gce_mocks.Wrapper)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.MockServerFromConfig(conf)
	defer srv.Close()

	srv.RunHelloWorld()

	// The GCE scheduler under test
	client := &gceClient{
		wrapper: wpr,
		project: "test-proj",
		zone:    "test-zone",
	}
	s := &Backend{conf, srv.Client, client}

	wpr.SetupMockInstanceTemplates()
	wpr.SetupMockMachineTypes()

	scheduler.ScheduleChunk(ctx, srv.DB, s, conf)
	workers := srv.GetWorkers()

	if len(workers) != 1 {
		t.Error("Expected a single worker")
		return
	}

	log.Debug("Workers", workers)
	w := workers[0]

	if w.Metadata["gce-template"] != "test-tpl" {
		t.Error("Worker has incorrect template")
	}

	wconf := conf
	wconf.Worker.ID = w.Id
	wconf.Worker.ServerAddress = conf.HostName + ":" + conf.RPCPort
	confYaml := string(wconf.ToYaml())
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
					DiskSizeGb: 14,
					DiskType:   "zones/test-zone/diskTypes/", // TODO??? this must be wrong
				},
			},
		},
		Name:        w.Id,
		MachineType: "zones/test-zone/machineTypes/test-mt",
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "funnel-config",
					Value: &confYaml,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"funnel"},
		},
	}
	wpr.On("InsertInstance", "test-proj", "test-zone", expected).Return(nil, nil)

	scheduler.Scale(ctx, srv.DB, s)
	wpr.AssertExpectations(t)
}

// Tests what happens when the InsertInstance() call fails the first couple times.
func TestInsertTempError(t *testing.T) {

	conf := basicConf()
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
	client.StartWorker("test-tpl", conf)
	client.StartWorker("test-tpl", conf)
	client.StartWorker("test-tpl", conf)
	wpr.AssertExpectations(t)
	// Clear the previous expected calls
	wpr.ExpectedCalls = nil
	wpr.SetupMockInstanceTemplates()
	wpr.SetupMockMachineTypes()

	// Now set InsertInstance to success
	confYaml := string(conf.ToYaml())
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
					DiskSizeGb: 14,
					DiskType:   "zones/test-zone/diskTypes/", // TODO??? this must be wrong
				},
			},
		},
		Name:        "test-worker",
		MachineType: "zones/test-zone/machineTypes/test-mt",
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "funnel-config",
					Value: &confYaml,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"funnel"},
		},
	}
	wpr.On("InsertInstance", "test-proj", "test-zone", expected).Return(nil, nil)

	client.StartWorker("test-tpl", conf)
	wpr.AssertExpectations(t)
}
