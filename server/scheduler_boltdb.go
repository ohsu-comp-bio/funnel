package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"time"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (taskBolt *TaskBolt) ReadQueue(n int) []*tes.Task {
	tasks := make([]*tes.Task, 0)
	taskBolt.db.View(func(tx *bolt.Tx) error {

		// Iterate over the TasksQueued bucket, reading the first `n` tasks
		c := tx.Bucket(TasksQueued).Cursor()
		for k, _ := c.First(); k != nil && len(tasks) < n; k, _ = c.Next() {
			id := string(k)
			task, _ := getTaskView(tx, id, tes.TaskView_FULL)
			tasks = append(tasks, task)
		}
		return nil
	})
	return tasks
}

// AssignTask assigns a task to a node. This updates the task state to Initializing,
// and updates the node (calls UpdateNode()).
func (taskBolt *TaskBolt) AssignTask(t *tes.Task, w *pbs.Node) error {
	return taskBolt.db.Update(func(tx *bolt.Tx) error {
		// TODO this is important! write a test for this line.
		//      when a task is assigned, its state is immediately Initializing
		//      even before the node has received it.
		err := transitionTaskState(tx, t.Id, tes.State_INITIALIZING)
		if err != nil {
			return err
		}
		taskIDBytes := []byte(t.Id)
		nodeIDBytes := []byte(w.Id)
		// TODO the database needs tests for this stuff. Getting errors during dev
		//      because it's easy to forget to link everything.
		key := append(nodeIDBytes, taskIDBytes...)
		err = tx.Bucket(NodeTasks).Put(key, taskIDBytes)
		if err != nil {
			return err
		}
		err = tx.Bucket(TaskNode).Put(taskIDBytes, nodeIDBytes)
		if err != nil {
			return err
		}
		return updateNode(tx, w)
	})
}

// UpdateNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (taskBolt *TaskBolt) UpdateNode(ctx context.Context, req *pbs.Node) (*pbs.UpdateNodeResponse, error) {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		return updateNode(tx, req)
	})
	resp := &pbs.UpdateNodeResponse{}
	return resp, err
}

func updateNode(tx *bolt.Tx, req *pbs.Node) error {
	// Get node
	node := getNode(tx, req.Id)

	if node.Version != 0 && req.Version != 0 && node.Version != req.Version {
		return errors.New("Version outdated")
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
		state := getTaskState(tx, id)

		// If the node has acknowledged that the task is complete,
		// unlink the task from the node.
		switch state {
		case Canceled, Complete, Error, SystemError:
			key := append([]byte(req.Id), []byte(id)...)
			tx.Bucket(NodeTasks).Delete(key)
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

	// TODO move to on-demand helper. i.e. don't store in DB
	updateAvailableResources(tx, node)
	node.Version = time.Now().Unix()
	return putNode(tx, node)
}

// TODO include active ports. maybe move Available out of the protobuf message
//      and expect this helper to be used?
func updateAvailableResources(tx *bolt.Tx, node *pbs.Node) {
	// Calculate available resources
	a := pbs.Resources{
		Cpus:   node.GetResources().GetCpus(),
		RamGb:  node.GetResources().GetRamGb(),
		DiskGb: node.GetResources().GetDiskGb(),
	}
	for _, taskID := range node.TaskIds {
		t, _ := getTaskView(tx, taskID, tes.TaskView_FULL)
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

// GetNode gets a node
func (taskBolt *TaskBolt) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	var node *pbs.Node
	var err error

	err = taskBolt.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(Nodes).Get([]byte(req.Id))
		if data == nil {
			return errNotFound
		}
		node = getNode(tx, req.Id)
		return nil
	})
	if err != nil {
		log.Debug("GetNode", "error", err, "nodeID", req.Id)
		if err == errNotFound {
			return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", err.Error(), req.Id))
		}
	}
	return node, err
}

// CheckNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (taskBolt *TaskBolt) CheckNodes() error {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(Nodes)
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &pbs.Node{}
			proto.Unmarshal(v, node)

			if node.State == pbs.NodeState_GONE {
				tx.Bucket(Nodes).Delete(k)
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

			if node.State == pbs.NodeState_UNINITIALIZED ||
				node.State == pbs.NodeState_INITIALIZING {

				// The node is initializing, which has a more liberal timeout.
				if d > taskBolt.conf.Scheduler.NodeInitTimeout {
					// Looks like the node failed to initialize. Mark it dead
					node.State = pbs.NodeState_DEAD
				}
			} else if d > taskBolt.conf.Scheduler.NodePingTimeout {
				// The node is stale/dead
				node.State = pbs.NodeState_DEAD
			} else {
				node.State = pbs.NodeState_ALIVE
			}
			// TODO when to delete nodes from the database?
			//      is dead node deletion an automatic garbage collection process?
			err := putNode(tx, node)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (taskBolt *TaskBolt) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	resp := &pbs.ListNodesResponse{}
	resp.Nodes = []*pbs.Node{}

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(Nodes)
		c := bucket.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			node := getNode(tx, string(k))
			resp.Nodes = append(resp.Nodes, node)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getNode(tx *bolt.Tx, id string) *pbs.Node {
	node := &pbs.Node{
		Id: id,
	}

	data := tx.Bucket(Nodes).Get([]byte(id))
	if data != nil {
		// TODO handle error
		proto.Unmarshal(data, node)
	}

	// Prefix scan for keys that start with node ID
	c := tx.Bucket(NodeTasks).Cursor()
	prefix := []byte(id)
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		taskID := string(v)
		state := getTaskState(tx, taskID)
		if tes.RunnableState(state) {
			node.TaskIds = append(node.TaskIds, taskID)
		}
	}

	if node.Metadata == nil {
		node.Metadata = map[string]string{}
	}

	return node
}

func putNode(tx *bolt.Tx, node *pbs.Node) error {
	// Tasks are not saved in the database under the node,
	// they are stored in a separate bucket and linked via an index.
	// The same protobuf message is used for both communication and database,
	// so we have to set nil here.
	//
	// Also, this modifies the node, so copy it first.
	p := proto.Clone(node).(*pbs.Node)
	p.TaskIds = nil
	data, _ := proto.Marshal(p)
	return tx.Bucket(Nodes).Put([]byte(p.Id), data)
}
