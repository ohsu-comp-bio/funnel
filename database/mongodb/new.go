package mongodb

import (
	"context"
	"fmt"

	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB provides an MongoDB database server backend.
type MongoDB struct {
	scheduler.UnimplementedSchedulerServiceServer
	client *mongo.Client
	conf   config.MongoDB
	active bool
}

func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	client, err := mongo.Connect(
		context.TODO(),
		options.Client().ApplyURI(conf.Addrs[0]))

	if err != nil {
		return nil, err
	}

	db := &MongoDB{
		client: client,
		conf:   conf,
		active: true,
	}
	return db, nil
}

func (db *MongoDB) tasks(client *mongo.Client) *mongo.Collection {
	return client.Database(db.conf.Database).Collection("tasks")
}

func (db *MongoDB) nodes(client *mongo.Client) *mongo.Collection {
	return client.Database(db.conf.Database).Collection("nodes")
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init() error {
	tasks := db.tasks(db.client)
	nodes := db.nodes(db.client)

	names, err := db.client.Database(db.conf.Database).ListCollectionNames(context.TODO(), bson.D{}, nil)
	if err != nil {
		return err
	}
	var tasksFound bool
	var nodesFound bool
	for _, name := range names {
		switch name {
		case "tasks":
			tasksFound = true
		case "nodes":
			nodesFound = true
		}
	}

	if !tasksFound {
		err = db.client.Database(db.conf.Database).CreateCollection(context.Background(), "tasks", nil)
		if err != nil {
			return fmt.Errorf("error creating tasks collection in database %s: %v", db.conf.Database, err)
		}

		indexModel := mongo.IndexModel{
			Keys: bson.D{
				{Key: "-id", Value: -1},
				{Key: "-creationtime", Value: -1},
			},
			Options: options.Index().SetUnique(true).SetSparse(true),
		}
		_, err = tasks.Indexes().CreateOne(context.TODO(), indexModel)
		if err != nil {
			return err
		}
	}

	if !nodesFound {
		err = db.client.Database(db.conf.Database).CreateCollection(context.Background(), "nodes", nil)
		if err != nil {
			return fmt.Errorf("error creating nodes collection in database %s: %v", db.conf.Database, err)
		}

		indexModel := mongo.IndexModel{
			Keys: bson.D{
				{Key: "-id", Value: -1},
			},
			Options: options.Index().SetUnique(true).SetSparse(true),
		}
		_, err = nodes.Indexes().CreateOne(context.TODO(), indexModel)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database session.
func (db *MongoDB) Close() {
	if db.active {
		db.client.Disconnect(context.TODO())
	}
	db.active = false
}
