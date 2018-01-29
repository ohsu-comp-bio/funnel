package boltdb

import (
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

// queueTask adds a task to the scheduler queue.
func (taskBolt *BoltDB) queueTask(task *tes.Task) error {
	taskID := task.Id
	idBytes := []byte(taskID)

	err := taskBolt.db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(TasksQueued).Put(idBytes, []byte{})
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't queue task: %s", err)
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

// PutNode put a node object into the database.
//
// For optimisic locking, if the node already exists and node.Version
// doesn't match the version in the database, an error is returned.
func (taskBolt *BoltDB) PutNode(ctx context.Context, node *pbs.Node) (*pbs.PutNodeResponse, error) {
	err := taskBolt.db.Update(func(tx *bolt.Tx) error {

		existing := &pbs.Node{}
		data := tx.Bucket(Nodes).Get([]byte(node.Id))
		if data != nil {
			proto.Unmarshal(data, existing)
		}

		if existing.GetVersion() != 0 && node.Version != existing.GetVersion() {
			return fmt.Errorf("Version outdated")
		}

		err := scheduler.UpdateNode(ctx, taskBolt, node, existing)
		if err != nil {
			return err
		}

		data, err = proto.Marshal(node)
		if err != nil {
			return err
		}
		return tx.Bucket(Nodes).Put([]byte(node.Id), data)
	})
	return &pbs.PutNodeResponse{}, err
}

// GetNode gets a node
func (taskBolt *BoltDB) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	var node *pbs.Node

	err := taskBolt.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(Nodes).Get([]byte(req.Id))
		if data == nil {
			return errNotFound
		}

		node = &pbs.Node{}
		return proto.Unmarshal(data, node)
	})

	if err == errNotFound {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", err.Error(), req.Id))
	}

	if err != nil {
		return nil, err
	}
	return node, nil
}

// DeleteNode deletes the given node.
// Currently, the node's version field is not checked.
func (taskBolt *BoltDB) DeleteNode(ctx context.Context, node *pbs.Node) (*pbs.DeleteNodeResponse, error) {
	// TODO we don't check version on delete. should we?
	return &pbs.DeleteNodeResponse{}, taskBolt.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(Nodes).Delete([]byte(node.Id))
	})
}

// ListNodes is an API endpoint that returns a list of nodes.
func (taskBolt *BoltDB) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	resp := &pbs.ListNodesResponse{}

	err := taskBolt.db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(Nodes)
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := &pbs.Node{}
			err := proto.Unmarshal(v, node)
			if err != nil {
				return err
			}
			resp.Nodes = append(resp.Nodes, node)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
