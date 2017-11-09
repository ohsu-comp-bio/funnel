package mongodb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// QueueTask adds a task to the scheduler queue.
func (db *MongoDB) QueueTask(task *tes.Task) error {
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *MongoDB) ReadQueue(n int) []*tes.Task {
	var tasks []*tes.Task
	err := db.tasks.Find(bson.M{"state": tes.State_QUEUED}).Select(basicView).Limit(n).All(&tasks)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return tasks
}

// PutNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (db *MongoDB) PutNode(ctx context.Context, node *pbs.Node) (*pbs.PutNodeResponse, error) {
	q := bson.M{"id": node.Id}

	if node.GetVersion() != 0 {
		q["version"] = node.GetVersion()
	}

	var existing pbs.Node
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

	return &pbs.PutNodeResponse{}, err
}

// GetNode gets a node
func (db *MongoDB) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	var node pbs.Node
	err := db.nodes.Find(bson.M{"id": req.Id}).One(&node)
	if err == mgo.ErrNotFound {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", mgo.ErrNotFound.Error(), req.Id))
	}
	return &node, err
}

// DeleteNode deletes a node
func (db *MongoDB) DeleteNode(ctx context.Context, req *pbs.Node) error {
	err := db.nodes.Remove(bson.M{"id": req.Id})
	if err == mgo.ErrNotFound {
		return grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", mgo.ErrNotFound.Error(), req.Id))
	}
	return err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *MongoDB) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	var nodes []*pbs.Node
	err := db.nodes.Find(nil).All(&nodes)
	if err != nil {
		return nil, err
	}
	out := &pbs.ListNodesResponse{
		Nodes: nodes,
	}
	return out, nil
}
