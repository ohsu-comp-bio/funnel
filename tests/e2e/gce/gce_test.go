package gce

import (
	"context"
	"github.com/go-test/deep"
	"github.com/ohsu-comp-bio/funnel/compute/gce"
	gcemock "github.com/ohsu-comp-bio/funnel/compute/gce/mocks"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/compute/v1"
	"testing"
	"time"
)

var log = logger.New("gce-e2e")

func init() {
	log.Configure(logger.DebugConfig())
}

type Funnel struct {
	*e2e.Funnel
	GCE               *gcemock.Wrapper
	InstancesInserted []*compute.Instance
}

func (f *Funnel) AddNode(id string, cpus uint32, ram, disk float64) {
	conf := f.Conf
	conf.Scheduler.Node.ID = id
	conf.Scheduler.Node.Metadata["gce"] = "yes"
	conf.Scheduler.Node.Resources.Cpus = cpus
	conf.Scheduler.Node.Resources.RamGb = ram
	conf.Scheduler.Node.Resources.DiskGb = disk

	n, err := scheduler.NewNode(conf)
	if err != nil {
		panic(err)
	}
	go n.Run(context.Background())
	time.Sleep(time.Second * 2)
}

func NewFunnel() *Funnel {
	conf := e2e.DefaultConfig()
	conf.Backend = "gce"

	// NOTE: matches hard-coded values in mock wrapper
	conf.Backends.GCE.Project = "test-proj"
	conf.Backends.GCE.Zone = "test-zone"

	backend, err := gce.NewMockBackend(conf)
	if err != nil {
		panic(err)
	}

	fun := &Funnel{
		Funnel: e2e.NewFunnel(conf),
		GCE:    backend.Wrapper,
	}

	backend.Wrapper.SetupMockMachineTypes()
	backend.Wrapper.SetupMockInstanceTemplates()

	// Set up the mock Google Cloud plugin so that it starts a local node.
	backend.Wrapper.On("InsertInstance", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			log.Debug("INSERT")
			opts := args[2].(*compute.Instance)
			fun.InstancesInserted = append(fun.InstancesInserted, opts)

			meta := &gce.Metadata{}
			meta.Instance.Name = opts.Name
			meta.Instance.Hostname = "localhost"

			for _, item := range opts.Metadata.Items {
				if item.Key == "funnel-node-serveraddress" {
					meta.Instance.Attributes.FunnelNodeServerAddress = *item.Value
				}
			}

			meta.Instance.Zone = conf.Backends.GCE.Zone
			meta.Project.ProjectID = conf.Backends.GCE.Project
			c, cerr := gce.WithMetadataConfig(conf, meta)

			if cerr != nil {
				panic(cerr)
			}

			n, err := scheduler.NewNode(c)
			if err != nil {
				panic(err)
			}
			go n.Run(context.Background())
		}).
		Return(nil, nil)

	fun.Scheduler = scheduler.NewScheduler(fun.SDB, backend, conf.Scheduler)
	fun.StartServer()

	return fun
}

func TestMultipleTasks(t *testing.T) {
	fun := NewFunnel()

	id := fun.Run(`
    --sh 'echo hello world'
  `)
	task1 := fun.Wait(id)

	id2 := fun.Run(`
    --sh 'echo hello world'
  `)
	task2 := fun.Wait(id2)

	if task1.State != tes.State_COMPLETE || task2.State != tes.State_COMPLETE {
		t.Fatal("Expected tasks to complete successfully")
	}

	// This test stems from a bug found during testing GCE node init.
	//
	// The problem was that the scheduler could schedule one task but not two,
	// because the Disk resources would first be reported by the GCE instance template,
	// but once the node sent an update, the resource information was incorrectly
	// reported and merged.
	resp, _ := fun.SDB.ListNodes(context.Background(), &pbs.ListNodesRequest{})
	if len(resp.Nodes) != 1 {
		t.Fatal("Expected one node")
	}
}

// Test that the correct information is being passed to the Google Cloud API
// during node creation.
func TestWrapper(t *testing.T) {
	fun := NewFunnel()

	// Run a task
	id := fun.Run(`
    --sh 'sleep 100'
  `)
	fun.WaitForRunning(id)
	defer fun.Cancel(id)

	// Check the node
	nodes := fun.ListNodes()
	log.Debug("Nodes", nodes)
	if len(nodes) != 1 {
		t.Error("Expected a single node")
		return
	}
	n := nodes[0]

	if n.Metadata["gce-template"] != "test-tpl" {
		t.Error("node has incorrect template")
	}

	addr := fun.Conf.Server.RPCAddress()
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
			Name:        n.Id,
			MachineType: "zones/test-zone/machineTypes/test-mt",
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "funnel-node-serveraddress",
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

// TestSchedToExisting tests the case where an existing node has capacity
// available for the task. In this case, there are no instance templates,
// so the scheduler will not create any new nodes.
func TestSchedToExisting(t *testing.T) {
	fun := NewFunnel()
	fun.AddNode("existing", 10, 100.0, 1000.0)

	// Run a task
	id := fun.Run(`
    --sh 'sleep 100'
  `)
	fun.WaitForRunning(id)
	defer fun.Cancel(id)

	nodes := fun.ListNodes()

	if len(nodes) != 1 {
		t.Error("Expected a single node")
	}

	log.Debug("Nodes", nodes)
	n := nodes[0]

	if n.Id != "existing" {
		t.Error("Task scheduled to unexpected node")
	}
}

// TestSchedStartNode tests the case where the scheduler wants to start a new
// GCE node instance from a instance template defined in the configuration.
// The scheduler calls the GCE API to get the template details and assigns
// a task to that unintialized node. The scaler then calls the GCE API to
// start the node.
func TestSchedStartNode(t *testing.T) {
	fun := NewFunnel()
	fun.AddNode("existing", 1, 100.0, 1000.0)

	id := fun.Run(`
    --sh 'sleep 100'
    --cpu 3
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	nodes := fun.ListNodes()

	if len(nodes) != 2 {
		log.Debug("Nodes", nodes)
		t.Error("Expected new node to be added to database")
		return
	}

	log.Debug("Nodes", nodes)
	if nodes[1].TaskIds[0] != id {
		t.Error("Expected node to have task ID")
	}
}

// TestPreferExistingNode tests the case where there is an existing node
// AND instance templates available. The existing node has capacity for the task,
// and the task should be scheduled to the existing node.
func TestPreferExistingNode(t *testing.T) {
	fun := NewFunnel()
	fun.AddNode("existing", 10, 100.0, 1000.0)

	id := fun.Run(`
    --sh 'sleep 100'
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	nodes := fun.ListNodes()

	if len(nodes) != 1 {
		t.Error("Expected no new nodes to be created")
	}

	expected := nodes[0]
	log.Debug("Nodes", nodes)

	if expected.Id != "existing" {
		t.Error("Task was scheduled to the wrong node")
	}
}

// Test submit multiple tasks at once when no nodes exist. Multiple nodes
// should be started.
func TestSchedStartMultipleNode(t *testing.T) {
	fun := NewFunnel()

	// NOTE: the machine type hard-coded in scheduler/gce/mocks/Wrapper_helpers.go
	//       has 3 CPUs.
	id1 := fun.Run(`
    --sh 'sleep 100'
    --cpu 2
  `)
	id2 := fun.Run(`
    --sh 'sleep 100'
    --cpu 2
  `)
	id3 := fun.Run(`
    --sh 'sleep 100'
    --cpu 2
  `)
	id4 := fun.Run(`
    --sh 'sleep 100'
    --cpu 2
  `)

	fun.WaitForRunning(id1, id2, id3, id4)
	defer fun.Cancel(id1)
	defer fun.Cancel(id2)
	defer fun.Cancel(id3)
	defer fun.Cancel(id4)
	nodes := fun.ListNodes()

	if len(nodes) != 4 {
		log.Debug("NODES", nodes)
		t.Error("Expected multiple nodes")
	}
}

// Test that assigning a task to a node correctly updates the available resources.
func TestUpdateAvailableResources(t *testing.T) {
	fun := NewFunnel()
	fun.AddNode("existing", 10, 100.0, 1000.0)

	id := fun.Run(`
    --sh 'sleep 100'
    --cpu 2
  `)

	fun.WaitForRunning(id)
	defer fun.Cancel(id)
	nodes := fun.ListNodes()

	if len(nodes) != 1 || nodes[0].Id != "existing" {
		log.Debug("NODES", nodes)
		t.Error("Expected a single, existing node")
	}

	if nodes[0].Available.Cpus != 8 {
		t.Error("Unexpected cpu count")
	}
}

// Try to reproduce a bug where available CPUs seems to overflow
func TestUpdateBugAvailableResources(t *testing.T) {
	fun := NewFunnel()
	fun.AddNode("existing-1", 8, 100.0, 1000.0)
	fun.AddNode("existing-2", 8, 100.0, 1000.0)

	id1 := fun.Run(`
    --sh 'sleep 100'
    --cpu 4
  `)
	id2 := fun.Run(`
    --sh 'sleep 100'
    --cpu 4
  `)
	id3 := fun.Run(`
    --sh 'sleep 100'
    --cpu 4
  `)

	fun.WaitForRunning(id1, id2, id3)
	defer fun.Cancel(id1)
	defer fun.Cancel(id2)
	defer fun.Cancel(id3)
	nodes := fun.ListNodes()

	log.Debug("NODES", nodes)

	if len(nodes) != 2 {
		t.Error("unexpected node count")
	}

	tot := nodes[0].Available.Cpus + nodes[1].Available.Cpus

	if tot != 4 {
		t.Error("Expected total available cpu count to be 4")
	}
}
