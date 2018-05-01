package builtin

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewNodeProcess returns a new NodeProcess instance.
func NewNodeProcess(conf config.Node, cli SchedulerServiceClient, factory Worker, log *logger.Logger) (*NodeProcess, error) {

	id := genID()
	log = log.WithFields("nodeID", id)
	log.Debug("NewNode", "config", conf)

	return &NodeProcess{
		conf: conf,
		log:  log,
		detail: &Node{
			Id:          id,
			State:       NodeState_ALIVE,
			Preemptible: conf.Preemptible,
			Zone:        conf.Zone,
			Hostname:    hostname(),
			Metadata:    conf.Metadata,
		},
		tasks:     map[string]*tes.Task{},
		workerRun: factory,
		stream:    &stream{client: cli},
	}, nil
}

// NodeProcess is a structure used for tracking available resources on a compute resource.
type NodeProcess struct {
	conf      config.Node
	log       *logger.Logger
	detail    *Node
	tasks     map[string]*tes.Task
	mtx       sync.Mutex
	waitGroup sync.WaitGroup
	workerRun Worker
	stream    *stream
}

// Run runs a node with the given config. This is responsible for communication
// with the server and starting task workers
func (n *NodeProcess) Run(ctx context.Context) error {

	defer n.finalize()

	go n.listen(ctx)

	for range util.Ticker(ctx, time.Duration(n.conf.UpdateRate)) {
		err := n.ping()
		if err != nil {
			n.log.Error("error pinging scheduler", "error", err)
		}
		if n.shouldShutdown() {
			break
		}
	}

	return nil
}

func (n *NodeProcess) shouldShutdown() bool {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	return len(n.tasks) == 0 && n.detail.State == NodeState_DRAIN
}

// finalize prepares the node for shutdown, disconnecting
// from the scheduler and stopping any running tasks.
func (n *NodeProcess) finalize() {
	// The node gets up to 10 seconds to finalize.
	timeout := time.After(10 * time.Second)

	wc := waitChan{}
	wc.Add(2)

	// Ping the scheduler with final state/details.
	n.detail.State = NodeState_GONE
	go func() {
		err := n.ping()
		if err != nil {
			n.log.Error("error sending final ping", "error", err)
		}
		wc.Done()
	}()

	// Stop any workers.
	go func() {
		n.waitGroup.Wait()
		wc.Done()
	}()

	select {
	case <-wc.Wait():
	case <-timeout:
	}

	n.stream.Close()
}

// Drain sets the node's state to DRAIN, which causes the node
// to stop accepting tasks.
func (n *NodeProcess) Drain() {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.detail.State = NodeState_DRAIN
}

func (n *NodeProcess) ping() error {

	if n.detail.Hostname == "" {
		n.detail.Hostname = hostname()
	}

	res, err := detectResources(n.conf)
	if err != nil {
		return fmt.Errorf("detecting resources: %s", err)
	}
	n.detail.Resources = &res
	n.detail.LastPing = time.Now().UnixNano()

	var tasks []*tes.Task
	var ids []string
	n.mtx.Lock()
	for id, task := range n.tasks {
		tasks = append(tasks, task)
		ids = append(ids, id)
	}
	n.mtx.Unlock()

	n.detail.TaskIds = ids
	n.detail.Available = availableResources(tasks, &res)

	err = n.stream.Send(n.detail)
	if err != nil {
		return fmt.Errorf("sending update: %s", err)
	}
	return nil
}

func (n *NodeProcess) listen(ctx context.Context) {
	for {
		control, err := n.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			n.log.Error("error receiving control", "error", err)
			return
		}
		// Avoid noisy "context canceled" logs.
		if status.Code(err) == codes.Canceled {
			continue
		}

		switch control.Type {
		case ControlType_CREATE_TASK:
			go n.runTask(ctx, control.Task)
		case ControlType_DRAIN_NODE:
			n.Drain()
		}
	}
}

func (n *NodeProcess) addTask(task *tes.Task) (added bool) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	_, exists := n.tasks[task.Id]
	if exists {
		return false
	}
	n.tasks[task.Id] = task
	return true
}

func (n *NodeProcess) removeTask(task *tes.Task) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	delete(n.tasks, task.Id)
}

func (n *NodeProcess) runTask(ctx context.Context, task *tes.Task) {

	if !n.addTask(task) {
		return
	}
	defer n.removeTask(task)

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
	// Enforce a minimum request of 1 cpu core
	if rcpus < 1 {
		rcpus = 1
	}
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
