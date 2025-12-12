package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
)

// WriteEvent creates an event for the server to handle.
func (db *PostgreSQL) WriteEvent(ctx context.Context, req *events.Event) error {
	r := util.Retrier{
		InitialInterval:     1 * time.Millisecond,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      5 * time.Minute,
		Multiplier:          1.5,
		RandomizationFactor: 0.5,
		MaxTries:            50,
		ShouldRetry: func(err error) bool {
			_, isTransitionError := err.(*tes.TransitionError)
			return !isTransitionError && err != tes.ErrNotFound && err != tes.ErrNotPermitted
		},
	}

	return r.Retry(ctx, func() error {
		return db.writeEvent(ctx, req)
	})
}

func (db *PostgreSQL) writeEvent(ctx context.Context, req *events.Event) error {
	// If this event creates a new task, insert it
	if req.Type == events.Type_TASK_CREATED {
		return db.createTask(ctx, req.GetTask())
	}

	// For all other events, update the existing task
	return db.updateTask(ctx, req)
}

func (db *PostgreSQL) createTask(ctx context.Context, task *tes.Task) error {
	data, err := proto.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshaling task: %v", err)
	}

	owner := server.GetUsername(ctx)
	query := `INSERT INTO tasks (id, data, state, owner) VALUES ($1, $2, $3, $4)`
	_, err = db.db.ExecContext(ctx, query, task.Id, data, task.State.String(), owner)
	if err != nil {
		return fmt.Errorf("inserting task: %v", err)
	}

	return nil
}

func (db *PostgreSQL) updateTask(ctx context.Context, req *events.Event) error {
	// Start a transaction
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Get the existing task
	var data []byte
	var owner string
	query := `SELECT data, owner FROM tasks WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, req.Id).Scan(&data, &owner)
	if err == sql.ErrNoRows {
		return tes.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("querying task: %v", err)
	}

	// Check permissions
	if !server.GetUser(ctx).IsAccessible(owner) {
		return tes.ErrNotPermitted
	}

	// Unmarshal task
	task := &tes.Task{}
	if err := proto.Unmarshal(data, task); err != nil {
		return fmt.Errorf("unmarshaling task: %v", err)
	}

	// Apply the event to the task
	if err := applyEvent(task, req); err != nil {
		return err
	}

	// Marshal updated task
	data, err = proto.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshaling updated task: %v", err)
	}

	// Update the task in the database
	updateQuery := `UPDATE tasks SET data = $1, state = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err = tx.ExecContext(ctx, updateQuery, data, task.State.String(), task.Id)
	if err != nil {
		return fmt.Errorf("updating task: %v", err)
	}

	return tx.Commit()
}

func applyEvent(task *tes.Task, req *events.Event) error {
	switch req.Type {
	case events.Type_TASK_STATE:
		from := task.State
		to := req.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		task.State = to

	case events.Type_TASK_START_TIME:
		task.GetTaskLog(0).StartTime = req.GetStartTime()

	case events.Type_TASK_END_TIME:
		task.GetTaskLog(0).EndTime = req.GetEndTime()

	case events.Type_TASK_OUTPUTS:
		task.GetTaskLog(0).Outputs = req.GetOutputs().Value

	case events.Type_TASK_METADATA:
		meta := req.GetMetadata().Value
		tl := task.GetTaskLog(0)
		if tl.Metadata == nil {
			tl.Metadata = map[string]string{}
		}
		for k, v := range meta {
			tl.Metadata[k] = v
		}

	case events.Type_EXECUTOR_START_TIME:
		task.GetExecLog(0, int(req.Index)).StartTime = req.GetStartTime()

	case events.Type_EXECUTOR_END_TIME:
		task.GetExecLog(0, int(req.Index)).EndTime = req.GetEndTime()

	case events.Type_EXECUTOR_EXIT_CODE:
		task.GetExecLog(0, int(req.Index)).ExitCode = req.GetExitCode()

	case events.Type_EXECUTOR_STDOUT:
		task.GetExecLog(0, int(req.Index)).Stdout = req.GetStdout()

	case events.Type_EXECUTOR_STDERR:
		task.GetExecLog(0, int(req.Index)).Stderr = req.GetStderr()

	case events.Type_SYSTEM_LOG:
		tl := task.GetTaskLog(0)
		tl.SystemLogs = append(tl.SystemLogs, req.SysLogString())
	}

	return nil
}
