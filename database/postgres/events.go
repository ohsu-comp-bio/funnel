package postgres

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// WriteEvent creates an event for the server to handle.
func (db *Postgres) WriteEvent(ctx context.Context, req *events.Event) error {
	// TODO: Implement

	return nil
}

func (db *Postgres) insertTask(ctx context.Context, task *tes.Task) error {
	// TODO: Implement

	return nil
}

func (db *Postgres) findTaskStateAndVersion(ctx context.Context, taskId string) (tes.State, interface{}, error) {
	// TODO: Implement

	return tes.State_UNKNOWN, nil, nil
}
