package postgres

import (
	"database/sql"
	"fmt"

	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	_ "github.com/lib/pq"
)

// PostgreSQL provides a PostgreSQL database server backend.
type PostgreSQL struct {
	scheduler.UnimplementedSchedulerServiceServer
	db   *sql.DB
	conf config.PostgreSQL
}

// NewPostgreSQL creates a new PostgreSQL database instance.
func NewPostgreSQL(conf config.PostgreSQL) (*PostgreSQL, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Database, conf.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %v", err)
	}

	// Set connection pool settings
	if conf.MaxOpenConns > 0 {
		db.SetMaxOpenConns(conf.MaxOpenConns)
	}
	if conf.MaxIdleConns > 0 {
		db.SetMaxIdleConns(conf.MaxIdleConns)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %v", err)
	}

	return &PostgreSQL{
		db:   db,
		conf: conf,
	}, nil
}

// Init creates the required database schema for PostgreSQL.
// This method must be called after NewPostgreSQL() to ensure proper
// database schema initialization before any database operations.
func (db *PostgreSQL) Init() error {
	// Create tasks table
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			data BYTEA NOT NULL,
			state TEXT NOT NULL,
			owner TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating tasks table: %v", err)
	}

	// Create indexes for better query performance
	_, err = db.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state);
		CREATE INDEX IF NOT EXISTS idx_tasks_owner ON tasks(owner);
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);
	`)
	if err != nil {
		return fmt.Errorf("creating indexes: %v", err)
	}

	// Create nodes table for scheduler support
	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			data BYTEA NOT NULL,
			last_ping TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating nodes table: %v", err)
	}

	return nil
}

// Close closes the database connection.
func (db *PostgreSQL) Close() {
	if db.db != nil {
		db.db.Close()
	}
}
