package scheduler

/*
TODO goals

- easy redeployment of upgraded node version
- no more node state concurrency issues
- faster node response time/scheduling
- correct processing of task queue
*/

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/util/rpc"
)

// NewNodeProcess returns a new NodeProcess instance.
func NewNodeProcess(conf config.Config, factory Worker, log *logger.Logger) (*NodeProcess, error) {
	log = log.WithFields("nodeID", conf.Node.ID)
	log.Debug("NewNode", "config", conf)

	return &NodeProcess{
		conf: conf,
		log:  log,
		detail: &Node{
			Id:          conf.Node.ID,
			State:       NodeState_ALIVE,
			Preemptible: conf.Node.Preemptible,
			Zone:        conf.Node.Zone,
			Hostname:    hostname(),
			Metadata:    conf.Node.Metadata,
		},
		workerRun: factory,
	}, nil
}

// NodeProcess is a structure used for tracking available resources on a compute resource.
type NodeProcess struct {
	conf      config.Config
	log       *logger.Logger
	detail    *Node
	tasks     sync.Map
	waitGroup sync.WaitGroup
	workerRun Worker
}

// Run runs a node with the given config. This is responsible for communication
// with the server and starting task workers
func (n *NodeProcess) Run(ctx context.Context) error {

	conn, err := rpc.Dial(ctx, n.conf.Server)
	if err != nil {
		return fmt.Errorf("connecting to server: %s", err)
	}
	defer conn.Close()

	client := NewSchedulerServiceClient(conn)
	stream, err := client.NodeChat(ctx)
	if err != nil {
		return fmt.Errorf("connecting to server: %s", err)
	}
	defer stream.CloseSend()

	go n.listen(ctx, stream)

	for range util.Ticker(ctx, time.Duration(n.conf.Node.UpdateRate)) {
		err := n.ping(ctx, stream)
		if err != nil {
			n.log.Error("error detecting resources", "error", err)
		}
	}

	// The workers get 10 seconds to finish up.
	timeout := time.After(10 * time.Second)
	done := make(chan struct{})
	go func() {
		n.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-timeout:
	}
	return nil
}

func (n *NodeProcess) ping(ctx context.Context, stream SchedulerService_NodeChatClient) error {

	if n.detail.Hostname == "" {
		n.detail.Hostname = hostname()
	}

	res, err := detectResources(n.conf.Node, n.conf.Worker.WorkDir)
	if err != nil {
		return fmt.Errorf("detecting resources: %s", err)
	}
	n.detail.Resources = &res
	n.detail.LastPing = time.Now().UnixNano()

	var tasks []*tes.Task
	n.tasks.Range(func(_, task interface{}) bool {
		tasks = append(tasks, task.(*tes.Task))
		return true
	})
	n.detail.Available = availableResources(tasks, &res)

	err = stream.Send(n.detail)
	if err != nil {
		return fmt.Errorf("sending update: %s", err)
	}
	return nil
}

func (n *NodeProcess) listen(ctx context.Context, stream SchedulerService_NodeChatClient) {
	for {
		task, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			n.log.Error("error receiving task", "error", err)
			return
		}
		go n.runTask(ctx, task)
	}
}

func (n *NodeProcess) runTask(ctx context.Context, task *tes.Task) {

	_, exists := n.tasks.LoadOrStore(task.Id, task)
	if exists {
		return
	}

	log := n.log.WithFields("ns", "worker", "taskID", task.Id)
	log.Info("Running task")

	n.waitGroup.Add(1)
	defer n.waitGroup.Done()

	defer func() {
		if r := recover(); r != nil {
			log.Error("caught panic task worker", r)
		}
	}()

	err := n.workerRun(ctx, task.Id)
	if err != nil {
		log.Error("error running task", err)
		return
	}

	log.Info("Task complete")
}

// availableResources calculates available resources given a list of tasks
// and base resources.
func availableResources(tasks []*tes.Task, res *Resources) *Resources {
	a := &Resources{
		Cpus:   res.GetCpus(),
		RamGb:  res.GetRamGb(),
		DiskGb: res.GetDiskGb(),
	}
	for _, t := range tasks {
		a = subtractResources(t, a)
	}
	return a
}

// subtractResources subtracts the resources requested by "task" from
// the node resources "in".
func subtractResources(t *tes.Task, in *Resources) *Resources {
	out := &Resources{
		Cpus:   in.GetCpus(),
		RamGb:  in.GetRamGb(),
		DiskGb: in.GetDiskGb(),
	}
	tres := t.GetResources()

	// Cpus are represented by an unsigned int, and if we blindly
	// subtract it will rollover to a very large number. So check first.
	rcpus := tres.GetCpuCores()
	if rcpus >= out.Cpus {
		out.Cpus = 0
	} else {
		out.Cpus -= rcpus
	}

	out.RamGb -= tres.GetRamGb()
	out.DiskGb -= tres.GetDiskGb()

	// Check minimum values.
	if out.Cpus < 0 {
		out.Cpus = 0
	}
	if out.RamGb < 0.0 {
		out.RamGb = 0.0
	}
	if out.DiskGb < 0.0 {
		out.DiskGb = 0.0
	}
	return out
}

func hostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	}
	return ""
}
