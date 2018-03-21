package mongodb

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *MongoDB) ReadQueue(n int) []*tes.Task {
	var tasks []*tes.Task
	err := db.tasks.Find(bson.M{"state": tes.State_QUEUED}).Sort("creationtime").Select(basicView).Limit(n).All(&tasks)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return tasks
}

// PutNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (db *MongoDB) PutNode(ctx context.Context, node *scheduler.Node) (*scheduler.PutNodeResponse, error) {
	q := bson.M{"id": node.Id}

	if node.GetVersion() != 0 {
		q["version"] = node.GetVersion()
	}

	var existing scheduler.Node
	err := db.nodes.Find(bson.M{"id": node.Id}).One(&existing)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}

	err = scheduler.UpdateNode(ctx, db, node, &existing)
	if err != nil {
		return nil, err
	}

	node.Version = node.GetVersion() + 1

	_, err = db.nodes.Upsert(q, node)

	return &scheduler.PutNodeResponse{}, err
}

// GetNode gets a node
func (db *MongoDB) GetNode(ctx context.Context, req *scheduler.GetNodeRequest) (*scheduler.Node, error) {
	var node scheduler.Node
	err := db.nodes.Find(bson.M{"id": req.Id}).One(&node)
	if err == mgo.ErrNotFound {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", mgo.ErrNotFound.Error(), req.Id))
	}
	return &node, err
}

// DeleteNode deletes a node
func (db *MongoDB) DeleteNode(ctx context.Context, req *scheduler.Node) (*scheduler.DeleteNodeResponse, error) {
	err := db.nodes.Remove(bson.M{"id": req.Id})
	if err == mgo.ErrNotFound {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", mgo.ErrNotFound.Error(), req.Id))
	}
	return nil, err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *MongoDB) ListNodes(ctx context.Context, req *scheduler.ListNodesRequest) (*scheduler.ListNodesResponse, error) {
	var nodes []*scheduler.Node
	err := db.nodes.Find(nil).All(&nodes)
	if err != nil {
		return nil, err
	}
	out := &scheduler.ListNodesResponse{
		Nodes: nodes,
	}
	return out, nil
}
