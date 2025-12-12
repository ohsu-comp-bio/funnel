package postgres

import (
	"context"
	"database/sql"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// GetTask gets a task by ID.
func (db *PostgreSQL) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var data []byte
	var owner string
	var state string

	query := `SELECT data, state, owner FROM tasks WHERE id = $1`
	err := db.db.QueryRowContext(ctx, query, req.Id).Scan(&data, &state, &owner)
	if err == sql.ErrNoRows {
		return nil, tes.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying task: %v", err)
	}

	// Check access permissions
	if !server.GetUser(ctx).IsAccessible(owner) {
		return nil, tes.ErrNotPermitted
	}

	// Unmarshal task data
	task := &tes.Task{}
	if err := proto.Unmarshal(data, task); err != nil {
		return nil, fmt.Errorf("unmarshaling task: %v", err)
	}

	// Apply the requested view
	switch req.View {
	case tes.View_MINIMAL.String():
		task = task.GetMinimalView()
	case tes.View_BASIC.String():
		task = task.GetBasicView()
	}

	return task, nil
}

// ListTasks returns a list of tasks.
func (db *PostgreSQL) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	pageSize := tes.GetPageSize(req.GetPageSize())
	var tasks []*tes.Task

	// Build query with filters
	query := `SELECT id, data, state, owner FROM tasks WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	// Filter by state
	if req.State != tes.Unknown {
		query += fmt.Sprintf(" AND state = $%d", argIdx)
		args = append(args, req.State.String())
		argIdx++
	}

	// Add pagination using ID comparison (consistent with MongoDB)
	if req.PageToken != "" {
		query += fmt.Sprintf(" AND id < $%d", argIdx)
		args = append(args, req.PageToken)
		argIdx++
	}

	// Order by id descending (most recent first, assuming IDs are monotonically increasing)
	query += " ORDER BY id DESC"

	// Limit results
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, pageSize+1) // Fetch one extra to determine if there's a next page

	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %v", err)
	}
	defer rows.Close()

	for rows.Next() && len(tasks) < pageSize {
		var id string
		var data []byte
		var state string
		var owner string

		if err := rows.Scan(&id, &data, &state, &owner); err != nil {
			return nil, fmt.Errorf("scanning task row: %v", err)
		}

		// Check access permissions
		if !server.GetUser(ctx).IsAccessible(owner) {
			continue
		}

		// Unmarshal task
		task := &tes.Task{}
		if err := proto.Unmarshal(data, task); err != nil {
			return nil, fmt.Errorf("unmarshaling task: %v", err)
		}

		// Filter by tags
		matchesTags := true
		for k, v := range req.GetTags() {
			tval, ok := task.Tags[k]
			if !ok || tval != v {
				matchesTags = false
				break
			}
		}
		if !matchesTags {
			continue
		}

		// Apply the requested view
		switch req.View {
		case tes.View_MINIMAL.String():
			task = task.GetMinimalView()
		case tes.View_BASIC.String():
			task = task.GetBasicView()
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating task rows: %v", err)
	}

	out := &tes.ListTasksResponse{
		Tasks: tasks,
	}

	// Set next page token if there are more results
	if len(tasks) == pageSize {
		// Check if there's actually a next page by seeing if we got the extra row
		var hasMore bool
		if rows.Next() {
			hasMore = true
		}
		if hasMore {
			out.NextPageToken = &tasks[len(tasks)-1].Id
		}
	}

	return out, nil
}
