package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
)

// Datastore provides a task database on Google Cloud Datastore.
type Datastore struct {
	client *datastore.Client
}

// NewDatastore returns a new Datastore instance with the given config.
func NewDatastore(conf config.Datastore) (*Datastore, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, conf.Project)
	if err != nil {
		return nil, err
	}
	return &Datastore{client}, nil
}
