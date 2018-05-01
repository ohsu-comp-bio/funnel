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
		workerRun: factory,
		client:    cli,
	}, nil
}

// NodeProcess is a structure used for tracking available resources on a compute resource.
type NodeProcess struct {
	conf      config.Node
	log       *logger.Logger
	detail    *Node
	tasks     sync.Map
	waitGroup sync.WaitGroup
	workerRun Worker
	client    SchedulerServiceClient
	stream    SchedulerService_NodeChatClient
}

// Run runs a node with the given config. This is responsible for communication
// with the server and starting task workers
func (n *NodeProcess) Run(ctx context.Context) error {

	n.connect(ctx)
	defer n.finalize()

	go n.listen(ctx)

	for range util.Ticker(ctx, time.Duration(n.conf.UpdateRate)) {
		err := n.ping()
		if err != nil {
			n.log.Error("error pinging scheduler", "error", err)
		}
	}

	return nil
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

	if n.stream != nil {
		n.stream.CloseSend()
	}
}

// connect tries hard to connect to the scheduler. It will retry until
// it connects or the context is canceled.
// TODO reconnect
func (n *NodeProcess) connect(ctx context.Context) {
	for range util.Ticker(ctx, time.Duration(n.conf.UpdateRate)) {
		// Closing the stream is managed by finalize(),
		// and since finalize() needs the stream after the context
		// has been canceled, we don't pass the parent context
		// to NodeChat() here.
		conn, err := n.client.NodeChat(context.Background())
		// Avoid noisy "context canceled" logs.
		if status.Code(err) == codes.Canceled {
			continue
		}
		if err != nil {
			n.log.Error("error connecting to server, will retry", "error", err)
			continue
		}
		n.stream = conn
		return
	}
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
	n.tasks.Range(func(_, rec interface{}) bool {
		task := rec.(*tes.Task)
		tasks = append(tasks, task)
		ids = append(ids, task.Id)
		return true
	})
	n.detail.TaskIds = ids
	n.detail.Available = availableResources(tasks, &res)

	if n.stream == nil {
		return fmt.Errorf("pinging server: no connection")
	}

	err = n.stream.Send(n.detail)
	// Avoid noisy "context canceled" logs.
	if status.Code(err) == codes.Canceled {
		return nil
	}
	if err != nil {
		return fmt.Errorf("sending update: %s", err)
	}
	return nil
}

// Drain sets the node's state to DRAIN, which causes the node
// to stop accepting tasks.
func (n *NodeProcess) Drain() {
	n.detail.State = NodeState_DRAIN
}

func (n *NodeProcess) listen(ctx context.Context) {
	// TODO figure out connect/reconnect
	if n.stream == nil {
		return
	}
	for {
		control, err := n.stream.Recv()
		if err == io.EOF {
			return
		}
		// Avoid noisy "context canceled" logs.
		if status.Code(err) == codes.Canceled {
			return
		}
		if err != nil {
			n.log.Error("error receiving control", "error", err)
			return
		}
		switch control.Type {
		case ControlType_CREATE_TASK:
			go n.runTask(ctx, control.Task)
		case ControlType_DRAIN_NODE:
			n.Drain()
		}
	}
}

func (n *NodeProcess) runTask(ctx context.Context, task *tes.Task) {

	_, exists := n.tasks.LoadOrStore(task.Id, task)
	if exists {
		return
	}
	defer n.tasks.Delete(task.Id)

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
