package dynamodb

import (
	"fmt"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

// QueueTask adds a task to the scheduler queue.
func (db *DynamoDB) QueueTask(task *tes.Task) error {
	return fmt.Errorf("QueueTask - Not Implemented")
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *DynamoDB) ReadQueue(n int) []*tes.Task {
	return nil
}

// PutNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (db *DynamoDB) PutNode(ctx context.Context, req *pbs.Node) (*pbs.PutNodeResponse, error) {
	return nil, fmt.Errorf("PutNode - Not Implemented")
}

// GetNode gets a node
func (db *DynamoDB) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	return nil, fmt.Errorf("GetNode - Not Implemented")
}

// DeleteNode deletes a node
func (db *DynamoDB) DeleteNode(ctx context.Context, req *pbs.Node) error {
	return fmt.Errorf("DeleteNode - Not Implemented")
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *DynamoDB) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	return nil, fmt.Errorf("ListNodes - Not Implemented")
}
