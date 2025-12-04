package postgres

import (
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *Postgres) ReadQueue(n int) []*tes.Task {
	// TODO: Implement

	return nil
}

// PutNode is an RPC endpoint that is used by nodes to send heartbeats
// and status updates, such as completed tasks. The server responds with updated
// information for the node, such as canceled tasks.
func (db *Postgres) PutNode(ctx context.Context, node *scheduler.Node) (*scheduler.PutNodeResponse, error) {
	// TODO: Implement

	return nil, nil
}

// GetNode gets a node
func (db *Postgres) GetNode(ctx context.Context, req *scheduler.GetNodeRequest) (*scheduler.Node, error) {
	// TODO: Implement

	return nil, nil
}

// DeleteNode deletes a node
func (db *Postgres) DeleteNode(ctx context.Context, req *scheduler.Node) (*scheduler.DeleteNodeResponse, error) {
	// TODO: Implement

	return nil, nil
}

// ListNodes is an API endpoint that returns a list of nodes.
func (db *Postgres) ListNodes(ctx context.Context, req *scheduler.ListNodesRequest) (*scheduler.ListNodesResponse, error) {
	// TODO: Implement

	return nil, nil
}
