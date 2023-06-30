package mongodb

import (
	"context"
	"fmt"
	"time"

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

// NewMongoDB returns a new MongoDB instance.
func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", conf.Addrs[0]),
	))
		
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
	// return sess.DB(db.conf.Database).C("tasks")
	return client.Database(db.conf.Database).Collection("tasks")
}

func (db *MongoDB) nodes(client *mongo.Client) *mongo.Collection {
	// return sess.DB(db.conf.Database).C("nodes")
	return client.Database(db.conf.Database).Collection("tasks")
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init() error {
	sess := db.client
	defer sess.Disconnect(context.TODO())
	tasks := db.tasks(sess)
	nodes := db.nodes(sess)
	res, err := tasks.InsertOne(context.Background(), bson.M{"hello": "world"})
	fmt.Println(res)

	names, err := sess.Database(db.conf.Database).ListCollectionNames(
		context.TODO(),
		bson.D{},
	)
	if err != nil {
		return fmt.Errorf("error listing collection names in database %s: %v", db.conf.Database, err)
	}
	var tasksFound bool
	var nodesFound bool
	for _, n := range names {
		switch n {
		case "tasks":
			tasksFound = true
		case "nodes":
			nodesFound = true
		}
	}

	if !tasksFound {
		err = db.client.Database(db.conf.Database).CreateCollection(context.TODO(), "tasks")
		if err != nil {
			return fmt.Errorf("error creating tasks collection in database %s: %v", db.conf.Database, err)
		}

		opts := options.Index()
		opts.SetUnique(true)
		opts.SetSparse(true)

		indexModel := mongo.IndexModel{
			// TODO: Map types such as bson.M are not valid.
			// See https://www.mongodb.com/docs/manual/indexes/#indexes for examples of valid documents.
			Keys: bson.M{
				"id": -1,
				"creationtime": -1,
			},
			Options: opts,
		}
		tasks.Indexes().CreateOne(context.TODO(), indexModel)

		if err != nil {
			return err
		}
	}

	if !nodesFound {
		err = db.client.Database(db.conf.Database).CreateCollection(context.TODO(), "nodes")
		if err != nil {
			return fmt.Errorf("error creating nodes collection in database %s: %v", db.conf.Database, err)
		}

		opts := options.Index()
		opts.SetUnique(true)
		opts.SetSparse(true)

		indexModel := mongo.IndexModel{
			Keys: "id",
			Options: opts,
		}
		nodes.Indexes().CreateOne(context.TODO(), indexModel)

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
