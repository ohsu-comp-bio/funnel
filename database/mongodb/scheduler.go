package mongodb

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/tes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *MongoDB) ReadQueue(n int) []*tes.Task {
	fmt.Println("Reading queue!")
	var tasks []*tes.Task
	opts := options.Find().SetSort(bson.M{"creationtime": 1}).SetLimit(int64(n))
	cursor, err := db.tasks(db.client).Find(context.TODO(), bson.M{"state": tes.State_QUEUED}, opts)

	err = cursor.All(context.TODO(), &tasks)
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
	nodes := db.nodes(db.client)

	q := bson.M{"id": node.Id}

	if node.GetVersion() != 0 {
		q["version"] = node.GetVersion()
	}

	var existing scheduler.Node
	err := nodes.FindOne(context.TODO(), bson.M{"id": node.Id}).Decode(&existing)
	if err != nil {
		return nil, err
	}

	db.GetTask(ctx, &tes.GetTaskRequest{Id: "foo"})
	err = scheduler.UpdateNode(ctx, db, node, &existing)
	if err != nil {
		return nil, err
	}

	node.Version = node.GetVersion() + 1

	opts := options.Update().SetUpsert(true)
	_, err = nodes.UpdateOne(context.TODO(), q, node, opts)

	return &scheduler.PutNodeResponse{}, err
}

// GetNode gets a node
func (db *MongoDB) GetNode(ctx context.Context, req *scheduler.GetNodeRequest) (*scheduler.Node, error) {
	var node scheduler.Node
	err := db.nodes(db.client).FindOne(context.TODO(), bson.M{"id": req.Id}).Decode(&node)
	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "%v: nodeID: %s", err, req.Id)
	}

	return &node, nil
}

// DeleteNode deletes a node
func (db *MongoDB) DeleteNode(ctx context.Context, req *scheduler.Node) (*scheduler.DeleteNodeResponse, error) {
	fmt.Println("DeleteNode", req.Id)
	_, err := db.nodes(db.client).DeleteOne(context.TODO(), bson.M{"id": req.Id})
	fmt.Println("DeleteNode", req.Id, err)
	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "%v: nodeID: %s", err, req.Id)
	}
	return nil, err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *MongoDB) ListNodes(ctx context.Context, req *scheduler.ListNodesRequest) (*scheduler.ListNodesResponse, error) {
	var nodes []*scheduler.Node
	cursor, err := db.nodes(db.client).Find(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &nodes)
	if err != nil {
		return nil, err
	}
	out := &scheduler.ListNodesResponse{
		Nodes: nodes,
	}

	return out, nil
}
