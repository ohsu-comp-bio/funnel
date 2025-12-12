package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PutNode is used by the scheduler to persist node information.
func (db *PostgreSQL) PutNode(ctx context.Context, node *scheduler.Node) (*scheduler.PutNodeResponse, error) {
	// Get existing node for version check
	var existing scheduler.Node
	var existingData []byte
	query := `SELECT data FROM nodes WHERE id = $1`
	err := db.db.QueryRowContext(ctx, query, node.Id).Scan(&existingData)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("querying existing node: %v", err)
	}
	
	if err == nil {
		if err := proto.Unmarshal(existingData, &existing); err != nil {
			return nil, fmt.Errorf("unmarshaling existing node: %v", err)
		}
		
		// Version check for optimistic locking
		if node.GetVersion() != 0 && node.GetVersion() != existing.GetVersion() {
			return nil, status.Error(codes.FailedPrecondition, "version mismatch")
		}
		
		// Update node using scheduler helper
		if err := scheduler.UpdateNode(ctx, db, node, &existing); err != nil {
			return nil, err
		}
	}
	
	// Increment version
	node.Version = node.GetVersion() + 1
	// Marshal and save node
	data, err := proto.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("marshaling node: %v", err)
	}

	upsertQuery := `
		INSERT INTO nodes (id, data, last_ping) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (id) 
		DO UPDATE SET data = $2, last_ping = CURRENT_TIMESTAMP
	`
	_, err = db.db.ExecContext(ctx, upsertQuery, node.Id, data)
	if err != nil {
		return nil, fmt.Errorf("upserting node: %v", err)
	}

	return &scheduler.PutNodeResponse{}, nil
}

// GetNode gets a node by ID.
func (db *PostgreSQL) GetNode(ctx context.Context, req *scheduler.GetNodeRequest) (*scheduler.Node, error) {
	var data []byte
	query := `SELECT data FROM nodes WHERE id = $1`
	err := db.db.QueryRowContext(ctx, query, req.Id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "nodeID: %s", req.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("querying node: %v", err)
	}

	node := &scheduler.Node{}
	if err := proto.Unmarshal(data, node); err != nil {
		return nil, fmt.Errorf("unmarshaling node: %v", err)
	}

	return node, nil
}

// ListNodes is used by the scheduler to get a list of nodes.
func (db *PostgreSQL) ListNodes(ctx context.Context, req *scheduler.ListNodesRequest) (*scheduler.ListNodesResponse, error) {
	query := `SELECT data FROM nodes ORDER BY id`
	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying nodes: %v", err)
	}
	defer rows.Close()

	var nodes []*scheduler.Node
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scanning node row: %v", err)
		}

		node := &scheduler.Node{}
		if err := proto.Unmarshal(data, node); err != nil {
			return nil, fmt.Errorf("unmarshaling node: %v", err)
		}

		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating node rows: %v", err)
	}

	return &scheduler.ListNodesResponse{Nodes: nodes}, nil
}

// DeleteNode is used by the scheduler to delete a node.
func (db *PostgreSQL) DeleteNode(ctx context.Context, req *scheduler.Node) (*scheduler.DeleteNodeResponse, error) {
	query := `DELETE FROM nodes WHERE id = $1`
	result, err := db.db.ExecContext(ctx, query, req.Id)
	if err != nil {
		return nil, fmt.Errorf("deleting node: %v", err)
	}
	
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, status.Errorf(codes.NotFound, "nodeID: %s", req.Id)
	}
	
	return &scheduler.DeleteNodeResponse{}, nil
}

// ReadQueue is used by the scheduler to get queued tasks from the database.
func (db *PostgreSQL) ReadQueue(n int) []*tes.Task {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT id, data FROM tasks 
		WHERE state = $1 
		ORDER BY created_at 
		LIMIT $2
	`

	rows, err := db.db.QueryContext(ctx, query, tes.State_QUEUED.String(), n)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tasks []*tes.Task
	for rows.Next() {
		var id string
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			continue
		}

		task := &tes.Task{}
		if err := proto.Unmarshal(data, task); err != nil {
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks
}


