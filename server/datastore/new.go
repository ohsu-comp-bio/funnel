package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
)

type KV interface {
	GetMulti()
}

type Datastore struct {
	client *datastore.Client
}

func NewDatastore(conf config.Datastore) (*Datastore, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, conf.Project)
	if err != nil {
		return nil, err
	}
	return &Datastore{client}, nil
}

func (d *Datastore) Close() error {
	return d.client.Close()
}
