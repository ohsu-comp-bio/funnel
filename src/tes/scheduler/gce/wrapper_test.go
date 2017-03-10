package gce

import (
	"errors"
	"github.com/stretchr/testify/mock"
	. "google.golang.org/api/compute/v1"
	"tes/logger"
	"tes/scheduler"
	gce_mocks "tes/scheduler/gce/mocks"
	server_mocks "tes/server/mocks"
	"testing"
)

func init() {
	logger.ForceColors()
}

// Test the scheduler while mocking out a lower level than what's in sched_test.go
// This mocks out the Wrapper interface, which allows better testing the logic
// in client.go and allows a more end-to-end test.
func TestWrapper(t *testing.T) {

	// Mock config
	conf := basicConf()
	// Set a different server address to test that it gets passed on to the worker
	conf.ServerAddress = "other:9090"
	// Add an instance template to the config. The scheduler uses these templates
	// to start new worker instances.
	conf.Schedulers.GCE.Templates = append(conf.Schedulers.GCE.Templates, "test-tpl")

	// Mock the GCE API wrapper
	wpr := new(gce_mocks.Wrapper)
	// Mock the server/database so we can easily control available workers
	srv := server_mocks.NewMockServer()
	defer srv.Close()

	srv.RunHelloWorld()

	// The GCE scheduler under test
	client := newClient(wpr)
	s := &gceScheduler{conf, srv.Client, client}

	wpr.On("ListMachineTypes", "test-proj", "test-zone").Return(&MachineTypeList{
		Items: []*MachineType{
			{
				Name:      "test-mt",
				GuestCpus: 3,
				MemoryMb:  12,
			},
		},
	}, nil)

	wpr.On("GetInstanceTemplate", "test-proj", "test-tpl").Return(&InstanceTemplate{
		Properties: &InstanceProperties{
			MachineType: "test-mt",
			Disks: []*AttachedDisk{
				{
					InitializeParams: &AttachedDiskInitializeParams{
						DiskSizeGb: 14,
					},
				},
			},
			Metadata: &Metadata{},
		},
	}, nil)

	scheduler.ScheduleChunk(srv.DB, s, conf)
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

	workerConf := conf.Worker
	workerConf.ID = w.Id
	workerConf.ServerAddress = conf.ServerAddress
	confYaml := string(workerConf.ToYaml())
	expected := &Instance{
		// TODO test that these fields get passed through from the template correctly.
		//      i.e. mock a more complex template
		CanIpForward:      false,
		CpuPlatform:       "",
		CreationTimestamp: "",
		Description:       "",
		Disks: []*AttachedDisk{
			{
				InitializeParams: &AttachedDiskInitializeParams{
					DiskSizeGb: 14,
					DiskType:   "zones/test-zone/diskTypes/", // TODO??? this must be wrong
				},
			},
		},
		Name:        w.Id,
		MachineType: "zones/test-zone/machineTypes/test-mt",
		Metadata: &Metadata{
			Items: []*MetadataItems{
				{
					Key:   "funnel-config",
					Value: &confYaml,
				},
			},
		},
	}
	wpr.On("InsertInstance", "test-proj", "test-zone", expected).Return(nil, nil)

	scheduler.Scale(srv.DB, s)
	wpr.AssertExpectations(t)
}

// Tests what happens when the InsertInstance() call fails the first couple times.
func TestInsertTempError(t *testing.T) {

	conf := basicConf().Worker
	conf.ID = "test-worker"
	wpr := new(gce_mocks.Wrapper)
	client := newClient(wpr)

	wpr.On("GetInstanceTemplate", "test-proj", "test-tpl").Return(&InstanceTemplate{
		Properties: &InstanceProperties{
			MachineType: "test-mt",
			Disks: []*AttachedDisk{
				{
					InitializeParams: &AttachedDiskInitializeParams{
						DiskSizeGb: 14,
					},
				},
			},
			Metadata: &Metadata{},
		},
	}, nil)

	// Set InsertInstance to return an error
	wpr.On("InsertInstance", "test-proj", "test-zone", mock.Anything).Return(nil, errors.New("TEST"))
	// Try to start the worker a few times
	// Do this a few times to exacerbate any errors.
	// e.g. a previous bug would build up a longer config string after every failure
	//      because cached data was being incorrectly shared.
	client.StartWorker("test-proj", "test-zone", "test-tpl", conf)
	client.StartWorker("test-proj", "test-zone", "test-tpl", conf)
	client.StartWorker("test-proj", "test-zone", "test-tpl", conf)
	wpr.AssertExpectations(t)

	// Now set InsertInstance to success
	confYaml := string(conf.ToYaml())
	expected := &Instance{
		// TODO test that these fields get passed through from the template correctly.
		//      i.e. mock a more complex template
		CanIpForward:      false,
		CpuPlatform:       "",
		CreationTimestamp: "",
		Description:       "",
		Disks: []*AttachedDisk{
			{
				InitializeParams: &AttachedDiskInitializeParams{
					DiskSizeGb: 14,
					DiskType:   "zones/test-zone/diskTypes/", // TODO??? this must be wrong
				},
			},
		},
		Name:        "test-worker",
		MachineType: "zones/test-zone/machineTypes/test-mt",
		Metadata: &Metadata{
			Items: []*MetadataItems{
				{
					Key:   "funnel-config",
					Value: &confYaml,
				},
			},
		},
	}
	// Clear the previous expected calls
	wpr.ExpectedCalls = nil
	wpr.On("InsertInstance", "test-proj", "test-zone", expected).Return(nil, nil)

	client.StartWorker("test-proj", "test-zone", "test-tpl", conf)
	wpr.AssertExpectations(t)
}
