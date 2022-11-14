package mongodb

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
)

// MongoDB provides an MongoDB database server backend.
type MongoDB struct {
	scheduler.UnimplementedSchedulerServiceServer
	sess   *mgo.Session
	conf   config.MongoDB
	active bool
}

// NewMongoDB returns a new MongoDB instance.
func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	sess, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:         conf.Addrs,
		Username:      conf.Username,
		Password:      conf.Password,
		Database:      conf.Database,
		Timeout:       time.Duration(conf.Timeout),
		AppName:       "funnel",
		PoolLimit:     4096,
		PoolTimeout:   0, // wait for connection to become available
		MinPoolSize:   10,
		MaxIdleTimeMS: 120000, // 2 min
	})
	if err != nil {
		return nil, err
	}
	db := &MongoDB{
		sess:   sess,
		conf:   conf,
		active: true,
	}
	return db, nil
}

func (db *MongoDB) tasks(sess *mgo.Session) *mgo.Collection {
	return sess.DB(db.conf.Database).C("tasks")
}

func (db *MongoDB) nodes(sess *mgo.Session) *mgo.Collection {
	return sess.DB(db.conf.Database).C("nodes")
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init() error {
	sess := db.sess.Copy()
	defer sess.Close()
	tasks := db.tasks(sess)
	nodes := db.nodes(sess)

	names, err := sess.DB(db.conf.Database).CollectionNames()
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
		err = tasks.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating tasks collection in database %s: %v", db.conf.Database, err)
		}

		err = tasks.EnsureIndex(mgo.Index{
			Key:        []string{"-id", "-creationtime"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		})
		if err != nil {
			return err
		}
	}

	if !nodesFound {
		err = nodes.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating nodes collection in database %s: %v", db.conf.Database, err)
		}

		err = nodes.EnsureIndex(mgo.Index{
			Key:        []string{"id"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database session.
func (db *MongoDB) Close() {
	if db.active {
		db.sess.Close()
	}
	db.active = false
}
