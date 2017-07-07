package gce

import (
	"context"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/scheduler/gce"
	gcemock "github.com/ohsu-comp-bio/funnel/scheduler/gce/mocks"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/compute/v1"
	"testing"
	"time"
)

var log = logger.New("gce-e2e")

type Funnel struct {
	*e2e.Funnel
	InstancesInserted []*compute.Instance
}

func (f *Funnel) AddWorker(id string, cpus uint32, ram, disk float64) {
	x := f.Conf.Worker
	x.ID = id
	x.Metadata["gce"] = "yes"
	x.Resources = config.Resources{
		Cpus:   cpus,
		RamGb:  ram,
		DiskGb: disk,
	}
	w, err := worker.NewWorker(x)
	if err != nil {
		panic(err)
	}
	go w.Run(context.Background())
	time.Sleep(time.Second * 2)
}

func NewFunnel() *Funnel {
	conf := e2e.DefaultConfig()

	// NOTE: matches hard-coded values in mock wrapper
	conf.Backends.GCE.Project = "test-proj"
	conf.Backends.GCE.Zone = "test-zone"

	backend, err := gce.NewMockBackend(conf)
	if err != nil {
		panic(err)
	}
	wrapper := new(gcemock.Wrapper)
	backend.SetWrapper(wrapper)

	fun := &Funnel{
		Funnel: e2e.NewFunnel(conf),
	}

	wrapper.SetupMockMachineTypes()
	wrapper.SetupMockInstanceTemplates()

	// Set up the mock Google Cloud plugin so that it starts a local worker.
	wrapper.On("InsertInstance", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			log.Debug("INSERT")
			opts := args[2].(*compute.Instance)
			fun.InstancesInserted = append(fun.InstancesInserted, opts)

			meta := &gce.Metadata{}
			meta.Instance.Name = opts.Name
			meta.Instance.Hostname = "localhost"

			for _, item := range opts.Metadata.Items {
				if item.Key == "funnel-worker-serveraddress" {
					meta.Instance.Attributes.FunnelWorkerServerAddress = *item.Value
				}
			}

			meta.Instance.Zone = conf.Backends.GCE.Zone
			meta.Project.ProjectID = conf.Backends.GCE.Project
			c, cerr := gce.WithMetadataConfig(conf, meta)

			if cerr != nil {
				panic(cerr)
			}

			w, err := worker.NewWorker(c.Worker)
			if err != nil {
				panic(err)
			}
			go w.Run(context.Background())
		}).
		Return(nil, nil)

	fun.Scheduler = scheduler.NewScheduler(fun.DB, backend, conf)
	fun.StartServer()

	return fun
}

func TestMultipleTasks(t *testing.T) {
	fun := NewFunnel()

	id := fun.Run(`
    --cmd 'echo hello world'
  `)
	task1 := fun.Wait(id)

	id2 := fun.Run(`
    --cmd 'echo hello world'
  `)
	task2 := fun.Wait(id2)

	if task1.State != tes.State_COMPLETE || task2.State != tes.State_COMPLETE {
		t.Fatal("Expected tasks to complete successfully")
	}

	// This test stems from a bug found during testing GCE worker init.
	//
	// The problem was that the scheduler could schedule one task but not two,
	// because the Disk resources would first be reported by the GCE instance template,
	// but once the worker sent an update, the resource information was incorrectly
	// reported and merged.
	resp, _ := fun.DB.ListWorkers(context.Background(), &pbf.ListWorkersRequest{})
	if len(resp.Workers) != 1 {
		t.Fatal("Expected one worker")
	}
}

// Test that the correct information is being passed to the Google Cloud API
// during worker creation.
func TestWrapper(t *testing.T) {
	fun := NewFunnel()

	// Run a task
	id := fun.Run(`
    --cmd 'sleep 100'
  `)
	fun.WaitForRunning(id)
	defer fun.Cancel(id)

	// Check the worker
	workers := fun.ListWorkers()
	log.Debug("Workers", workers)
	if len(workers) != 1 {
		t.Error("Expected a single worker")
		return
	}
	w := workers[0]

	if w.Metadata["gce-template"] != "test-tpl" {
		t.Error("Worker has incorrect template")
	}

	addr := fun.Conf.RPCAddress()
	d := deep.Equal(fun.InstancesInserted, []*compute.Instance{
		{
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
			Name:        w.Id,
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
		},
	})
	if d != nil {
		t.Fatal("unexpected instances inserted", d)
	}
}

// TestSchedToExisting tests the case where an existing worker has capacity
// available for the task. In this case, there are no instance templates,
// so the scheduler will not create any new workers.
func TestSchedToExisting(t *testing.T) {
	fun := NewFunnel()
	fun.AddWorker("existing", 10, 100.0, 1000.0)

	// Run a task
	id := fun.Run(`
    --cmd 'sleep 100'
  `)
	fun.WaitForRunning(id)
	defer fun.Cancel(id)

	workers := fun.ListWorkers()

	if len(workers) != 1 {
		t.Error("Expected a single worker")
	}

	log.Debug("Workers", workers)
	w := workers[0]

	if w.Id != "existing" {
		t.Error("Task scheduled to unexpected worker")
	}
}

// TestSchedStartWorker tests the case where the scheduler wants to start a new
// GCE worker instance from a instance template defined in the configuration.
// The scheduler calls the GCE API to get the template details and assigns
// a task to that unintialized worker. The scaler then calls the GCE API to
// start the worker.
func TestSchedStartWorker(t *testing.T) {
	fun := NewFunnel()
	fun.AddWorker("existing", 1, 100.0, 1000.0)

	id := fun.Run(`
    --cmd 'sleep 100'
    --cpu 3
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	workers := fun.ListWorkers()

	if len(workers) != 2 {
		log.Debug("Workers", workers)
		t.Error("Expected new worker to be added to database")
		return
	}

	log.Debug("Workers", workers)
	if workers[1].TaskIds[0] != id {
		t.Error("Expected worker to have task ID")
	}
}

// TestPreferExistingWorker tests the case where there is an existing worker
// AND instance templates available. The existing worker has capacity for the task,
// and the task should be scheduled to the existing worker.
func TestPreferExistingWorker(t *testing.T) {
	fun := NewFunnel()
	fun.AddWorker("existing", 10, 100.0, 1000.0)

	id := fun.Run(`
    --cmd 'sleep 100'
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	workers := fun.ListWorkers()

	if len(workers) != 1 {
		t.Error("Expected no new workers to be created")
	}

	expected := workers[0]
	log.Debug("Workers", workers)

	if expected.Id != "existing" {
		t.Error("Task was scheduled to the wrong worker")
	}
}

// Test submit multiple tasks at once when no workers exist. Multiple workers
// should be started.
func TestSchedStartMultipleWorker(t *testing.T) {
	fun := NewFunnel()

	// NOTE: the machine type hard-coded in scheduler/gce/mocks/Wrapper_helpers.go
	//       has 3 CPUs.
	id1 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 2
  `)
	id2 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 2
  `)
	id3 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 2
  `)
	id4 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 2
  `)

	fun.WaitForRunning(id1, id2, id3, id4)
	defer fun.Cancel(id1)
	defer fun.Cancel(id2)
	defer fun.Cancel(id3)
	defer fun.Cancel(id4)
	workers := fun.ListWorkers()

	if len(workers) != 4 {
		log.Debug("WORKERS", workers)
		t.Error("Expected multiple workers")
	}
}

// Test that assigning a task to a worker correctly updates the available resources.
func TestUpdateAvailableResources(t *testing.T) {
	fun := NewFunnel()
	fun.AddWorker("existing", 10, 100.0, 1000.0)

	id := fun.Run(`
    --cmd 'sleep 100'
    --cpu 2
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	workers := fun.ListWorkers()

	if len(workers) != 1 || workers[0].Id != "existing" {
		log.Debug("WORKERS", workers)
		t.Error("Expected a single, existing worker")
	}

	if workers[0].Available.Cpus != 8 {
		t.Error("Unexpected cpu count")
	}
}

// Try to reproduce a bug where available CPUs seems to overflow
func TestUpdateBugAvailableResources(t *testing.T) {
	fun := NewFunnel()
	fun.AddWorker("existing-1", 8, 100.0, 1000.0)
	fun.AddWorker("existing-2", 8, 100.0, 1000.0)

	id1 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 4
  `)
	id2 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 4
  `)
	id3 := fun.Run(`
    --cmd 'sleep 100'
    --cpu 4
  `)

	fun.WaitForRunning(id1, id2, id3)
	defer fun.Cancel(id1)
	defer fun.Cancel(id2)
	defer fun.Cancel(id3)
	workers := fun.ListWorkers()

	log.Debug("WORKERS", workers)

	if len(workers) != 2 {
		t.Error("unexpected worker count")
	}

	tot := workers[0].Available.Cpus + workers[1].Available.Cpus

	if tot != 4 {
		t.Error("Expected total available cpu count to be 4")
	}
}
