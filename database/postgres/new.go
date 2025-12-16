package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func getConnString(conf config.Postgres, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		conf.User,
		conf.Password,
		conf.Host,
		db,
	)
}

func getDefaultConnString(conf config.Postgres) string {
	return getConnString(conf, conf.Database)
}

func ensureDatabaseExists(ctx context.Context, conf *config.Postgres) error {
	connStr := getConnString(*conf, "postgres")

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to administrative database 'postgres': %w", err)
	}
	defer conn.Close(ctx)

	createDBSQL := fmt.Sprintf("CREATE DATABASE %s OWNER %s", conf.Database, conf.User)

	_, err = conn.Exec(ctx, createDBSQL)
	if err != nil {
		checkDBSQL := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", conf.Database)
		var exists int
		err = conn.QueryRow(ctx, checkDBSQL).Scan(&exists)

		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("failed to query database existence: %w", err)
		}

		if err == pgx.ErrNoRows {
			_, err = conn.Exec(ctx, createDBSQL)
			if err != nil {
				return fmt.Errorf("failed to create database '%s' using SQL '%s': %w", conf.Database, createDBSQL, err)
			}
		}
	}

	grantSchemaSQL := "GRANT CREATE ON SCHEMA public TO " + conf.User
	if _, err := conn.Exec(ctx, grantSchemaSQL); err != nil {
		return fmt.Errorf("failed to grant schema permissions: %w", err)
	}

	return nil
}

func NewPostgres(conf *config.Postgres) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout.GetDuration().AsDuration())
	defer cancel()

	err := ensureDatabaseExists(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("error connecting to or creating database resources: %w", err)
	}

	connString := getDefaultConnString(*conf)

	poolConf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Postgres connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
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

	// Tasks
	tasks := `
    CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(255) PRIMARY KEY,
		state VARCHAR(50) NOT NULL,
		owner VARCHAR(255),
		creation_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		version BIGINT DEFAULT 0 NOT NULL,
		data JSONB
    );
    `

	if _, err := db.client.Exec(ctx, tasks); err != nil {
		return fmt.Errorf("failed to create 'tasks' table: %w", err)
	}

	// Nodes
	nodes := `
    CREATE TABLE IF NOT EXISTS nodes (
		id VARCHAR(255) PRIMARY KEY,
		state VARCHAR(50) NOT NULL,
		owner VARCHAR(255),
		version BIGINT DEFAULT 0 NOT NULL,
		last_heartbeat TIMESTAMP WITH TIME ZONE,
		data JSONB
    );
    `

	if _, err := db.client.Exec(ctx, nodes); err != nil {
		return fmt.Errorf("failed to create 'nodes' table: %w", err)
	}

	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks (state);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_owner ON tasks (owner);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_creation_time ON tasks (creation_time DESC);",
		"CREATE INDEX IF NOT EXISTS idx_nodes_state ON nodes (state);",
		"CREATE INDEX IF NOT EXISTS idx_nodes_owner ON nodes (owner);",
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
