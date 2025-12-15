package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestNewPostgres(t *testing.T) {
	conf := &config.Postgres{
		Host:     "localhost:5432",
		Database: "funnel",
		User:     "funnel",
		Password: "example",

		Timeout: &config.TimeoutConfig{
			TimeoutOption: &config.TimeoutConfig_Duration{
				Duration: durationpb.New(5 * time.Second),
			},
		},
	}

	t.Logf("Starting Postgres Container...")
	connStr, container, err := newTestPostgres(conf)
	if err != nil {
		t.Fatalf("Failed to start test PostgreSQL container: %v", err)
	}

	defer func() {
		t.Logf("Terminating Postgres Container...")
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	t.Logf("Container ID: %s", container.GetContainerID())
	t.Logf("Connecting to Postgres: %s", connStr)

	db, err := NewPostgres(conf)
	if err != nil {
		t.Fatalf("Failed to create Postgres instance: %v", err)
	}
	defer db.client.Close()

	if !db.active {
		t.Errorf("Expected database to be active")
	}

	if err := db.Init(); err != nil {
		t.Fatalf("error creating database resources: %v", err)
	}
}

func newTestPostgres(conf *config.Postgres) (string, testcontainers.Container, error) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(conf.Database),
		postgres.WithUsername(conf.User),
		postgres.WithPassword(conf.Password),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	conf.Host = "localhost:" + port.Port()

	// Construct the connection string
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s",
		conf.User,
		conf.Password,
		port.Port(),
		conf.Database)
	return connStr, postgresContainer, nil
}
