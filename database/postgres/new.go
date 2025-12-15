package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
)

// Postgres provides a PostgreSQL database server backend.
type Postgres struct {
	scheduler.UnimplementedSchedulerServiceServer
	client *pgxpool.Pool // Use a connection pool for managing multiple connections
	conf   config.Postgres
	active bool
}

func getConnectionString(conf config.Postgres) string {
	// Build the PostgreSQL connection string
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Database,
	)
}

func NewPostgres(conf *config.Postgres) (*Postgres, error) {
	// Create a connection pool
	poolConfig, err := pgxpool.ParseConfig(getConnectionString(*conf))
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		client: pool,
		conf:   *conf,
		active: true,
	}, nil
}

func (db *Postgres) context() (context.Context, context.CancelFunc) {
	// Use a timeout from the configuration
	return context.WithTimeout(context.Background(), db.conf.Timeout.GetDuration().AsDuration())
}

func (db *Postgres) wrap(ctx context.Context) (context.Context, context.CancelFunc) {
	// TODO: Implement

	return nil, nil
}

func (db *Postgres) tasks() {
	// TODO: Implement
}

// Init creates required tables and indexes in Postgres
func (db *Postgres) Init() error {
	ctx, cancel := db.context()
	defer cancel()

	tasks := `
    CREATE TABLE IF NOT EXISTS tasks (
        id VARCHAR(255) PRIMARY KEY,
        state VARCHAR(50) NOT NULL,
        user_id VARCHAR(255),
        creation_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        -- Store the full task object as JSONB for flexible schema updates
        data JSONB
    );
    `

	if _, err := db.client.Exec(ctx, tasks); err != nil {
		return fmt.Errorf("failed to create 'tasks' table: %w", err)
	}

	// TODO: Add Node Table?

	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks (state);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_creation_time ON tasks (creation_time DESC);",
		"CREATE INDEX IF NOT EXISTS idx_nodes_state ON nodes (state);",
	}

	for _, query := range indices {
		if _, err := db.client.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create index with query '%s': %w", query, err)
		}
	}

	return nil
}

// Close closes the database session.
func (db *Postgres) Close() {
	// TODO: Implement
}
