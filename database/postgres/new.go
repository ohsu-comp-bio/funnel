package postgres

import (
	"context"

	"github.com/jackc/pgx"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
)

// MongoDB provides an MongoDB database server backend.
type Postgres struct {
	scheduler.UnimplementedSchedulerServiceServer
	client   *pgx.Conn
	database *pgx.Database
	conf     config.Postgres
	active   bool
}

func NewPostgres(conf config.Postgres) (*Postgres, error) {
	// TODO: Implement

	return &Postgres{}, nil
}

func (db *Postgres) context() (context.Context, context.CancelFunc) {
	// TODO: Implement

	return db.wrap(context.Background())
}

func (db *Postgres) wrap(ctx context.Context) (context.Context, context.CancelFunc) {
	// TODO: Implement

	return nil, nil
}

func (db *Postgres) collection(name string) {
	// TODO: Implement
}

func (db *Postgres) nodes() {
	// TODO: Implement
}

func (db *Postgres) tasks() {
	// TODO: Implement
}

// Init creates tables in MongoDB.
func (db *Postgres) Init() error {
	// TODO: Implement

	return nil
}

// Close closes the database session.
func (db *Postgres) Close() {
	// TODO: Implement
}
