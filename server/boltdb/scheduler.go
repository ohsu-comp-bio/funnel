package boltdb

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// QueueTask adds a task to the scheduler queue.
func (taskBolt *BoltDB) QueueTask(task *tes.Task) error {
	taskID := task.Id
	idBytes := []byte(taskID)

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(TasksQueued).Put(idBytes, []byte{})
		return nil
	})
	if err != nil {
		log.Error("Error queuing task", err)
		return err
	}
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (taskBolt *BoltDB) ReadQueue(n int) []*tes.Task {
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
func (taskBolt *BoltDB) AssignTask(t *tes.Task, w *pbs.Node) error {
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

		return taskBolt.updateNode(tx, w)
	})
}

func (taskBolt *BoltDB) updateNode(tx *bolt.Tx, req *pbs.Node) error {
	ctx := context.Background()
	node := getNode(tx, req.Id)

	terminalTaskIDs, err := scheduler.UpdateNode(ctx, taskBolt, node, req)
	if err != nil {
		return err
	}
	for _, id := range terminalTaskIDs {
		key := append([]byte(req.Id), []byte(id)...)
		tx.Bucket(NodeTasks).Delete(key)
	}
	return putNode(tx, node)
}

// UpdateNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (taskBolt *BoltDB) UpdateNode(ctx context.Context, req *pbs.Node) (*pbs.UpdateNodeResponse, error) {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		return taskBolt.updateNode(tx, req)
	})
	resp := &pbs.UpdateNodeResponse{}
	return resp, err
}

// GetNode gets a node
func (taskBolt *BoltDB) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
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
func (taskBolt *BoltDB) CheckNodes() error {
	var nodes []*pbs.Node
	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(Nodes)
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &pbs.Node{}
			proto.Unmarshal(v, node)
			nodes = append(nodes, node)
		}
		return nil
	})

	if err != nil {
		return err
	}

	updated := scheduler.UpdateNodeState(nodes, taskBolt.conf.Scheduler)
	if updated == nil {
		return nil
	}

	return taskBolt.db.Update(func(tx *bolt.Tx) error {
		for _, node := range updated {
			if node.State == pbs.NodeState_GONE {
				tx.Bucket(Nodes).Delete([]byte(node.Id))
			} else {
				err := putNode(tx, node)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ListNodes is an API endpoint that returns a list of nodes.
func (taskBolt *BoltDB) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
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
