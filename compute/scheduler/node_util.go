package scheduler

import (
	"context"
	"errors"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// TODO include active ports. maybe move Available out of the protobuf message
//      and expect this helper to be used?
func UpdateAvailableResources(ctx context.Context, tasks tes.TaskServiceServer, node *pbs.Node) {
	// Calculate available resources
	a := pbs.Resources{
		Cpus:   node.GetResources().GetCpus(),
		RamGb:  node.GetResources().GetRamGb(),
		DiskGb: node.GetResources().GetDiskGb(),
	}
	for _, taskID := range node.TaskIds {
		t, _ := tasks.GetTask(ctx, &tes.GetTaskRequest{
			Id:   taskID,
			View: tes.TaskView_FULL,
		})

		res := t.GetResources()

		// Cpus are represented by an unsigned int, and if we blindly
		// subtract it will rollover to a very large number. So check first.
		rcpus := res.GetCpuCores()
		if rcpus >= a.Cpus {
			a.Cpus = 0
		} else {
			a.Cpus -= rcpus
		}

		a.RamGb -= res.GetRamGb()
		a.DiskGb -= res.GetSizeGb()

		if a.Cpus < 0 {
			a.Cpus = 0
		}
		if a.RamGb < 0.0 {
			a.RamGb = 0.0
		}
		if a.DiskGb < 0.0 {
			a.DiskGb = 0.0
		}
	}
	node.Available = &a
}

func UpdateNode(ctx context.Context, tasks tes.TaskServiceServer, node *pbs.Node, req *pbs.Node) ([]string, error) {
	var terminalTaskIDs []string

	if node.Version != 0 && req.Version != 0 && node.Version != req.Version {
		return nil, errors.New("Version outdated")
	}

	node.LastPing = time.Now().Unix()
	node.State = req.GetState()

	if req.Resources != nil {
		if node.Resources == nil {
			node.Resources = &pbs.Resources{}
		}
		// Merge resources
		if req.Resources.Cpus > 0 {
			node.Resources.Cpus = req.Resources.Cpus
		}
		if req.Resources.RamGb > 0 {
			node.Resources.RamGb = req.Resources.RamGb
		}
	}

	// update disk usage while idle
	if len(req.TaskIds) == 0 {
		if req.GetResources().GetDiskGb() > 0 {
			node.Resources.DiskGb = req.Resources.DiskGb
		}
	}

	// Reconcile node's task states with database
	for _, id := range req.TaskIds {
		task, _ := tasks.GetTask(ctx, &tes.GetTaskRequest{
			Id:   id,
			View: tes.TaskView_MINIMAL,
		})
		state := task.GetState()

		// If the node has acknowledged that the task is complete,
		// unlink the task from the node.
		switch state {
		case tes.State_CANCELED, tes.State_COMPLETE, tes.State_ERROR, tes.State_SYSTEM_ERROR:
			terminalTaskIDs = append(terminalTaskIDs, id)
			// update disk usage once a task completes
			if req.GetResources().GetDiskGb() > 0 {
				node.Resources.DiskGb = req.Resources.DiskGb
			}
		}
	}

	if node.Metadata == nil {
		node.Metadata = map[string]string{}
	}
	for k, v := range req.Metadata {
		node.Metadata[k] = v
	}

	UpdateAvailableResources(ctx, tasks, node)
	node.Version = time.Now().Unix()
	return terminalTaskIDs, nil
}

func UpdateNodeState(nodes []*pbs.Node, conf config.Scheduler) []*pbs.Node {
	var updated []*pbs.Node
	for _, node := range nodes {
		prevState := node.State

		if node.State == pbs.NodeState_GONE {
			updated = append(updated, node)
			continue
		}

		if node.LastPing == 0 {
			// This shouldn't be happening, because nodes should be
			// created with LastPing, but give it the benefit of the doubt
			// and leave it alone.
			continue
		}

		lastPing := time.Unix(node.LastPing, 0)
		d := time.Since(lastPing)

		if node.State == pbs.NodeState_UNINITIALIZED || node.State == pbs.NodeState_INITIALIZING {

			// The node is initializing, which has a more liberal timeout.
			if d > conf.NodeInitTimeout {
				// Looks like the node failed to initialize. Mark it dead
				node.State = pbs.NodeState_DEAD
			}

		} else if node.State == pbs.NodeState_DEAD && d > conf.NodeDeadTimeout {
			// The node has been dead for long enough.
			node.State = pbs.NodeState_GONE

		} else if d > conf.NodePingTimeout {
			// The node hasn't pinged in awhile, mark it dead.
			node.State = pbs.NodeState_DEAD

		} else {
			node.State = pbs.NodeState_ALIVE
		}

		if prevState != node.State {
			updated = append(updated, node)
		}
	}
	return updated
}
