package dynamodb

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

// QueueTask adds a task to the scheduler queue.
func (db *DynamoDB) QueueTask(task *tes.Task) error {
	log.Error("QueueTask - Not Implemented")
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *DynamoDB) ReadQueue(n int) []*tes.Task {
	log.Error("ReadQueue - Not Implemented")
	return nil
}

// AssignTask assigns a task to a node. This updates the task state to Initializing,
// and updates the node (calls PutNode()).
func (db *DynamoDB) AssignTask(t *tes.Task, w *pbs.Node) error {
	log.Error("AssignTask - Not Implemented")
	return nil
}

// PutNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (db *DynamoDB) PutNode(ctx context.Context, req *pbs.Node) (*pbs.PutNodeResponse, error) {
	log.Error("PutNode - Not Implemented")
	return nil, nil
}

// GetNode gets a node
func (db *DynamoDB) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	log.Error("GetNodes - Not Implemented")
	return nil, nil
}

// CheckNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (db *DynamoDB) CheckNodes() error {
	log.Error("CheckNodes - Not Implemented")
	return nil
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *DynamoDB) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	log.Error("ListNodes - Not Implemented")
	return nil, nil
}
