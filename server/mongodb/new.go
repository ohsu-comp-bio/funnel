package mongodb

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/config"
	mgo "gopkg.in/mgo.v2"
)

// MongoDB provides an MongoDB database server backend.
type MongoDB struct {
	sess  *mgo.Session
	conf  config.MongoDB
	tasks *mgo.Collection
	nodes *mgo.Collection
	// events  *mgo.Collection
}

// NewMongoDB returns a new MongoDB instance.
func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	sess, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    conf.Addrs,
		Username: conf.Username,
		Password: conf.Password,
		Database: conf.Database,
		Timeout:  conf.Timeout,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		return nil, err
	}
	db := &MongoDB{
		sess:  sess,
		conf:  conf,
		tasks: sess.DB(conf.Database).C("tasks"),
		nodes: sess.DB(conf.Database).C("nodes"),
	}
	if err := db.init(); err != nil {
		return nil, err
	}
	return db, nil
}

// init creates tables in MongoDB.
func (db *MongoDB) init() error {
	names, err := db.sess.DB(db.conf.Database).CollectionNames()
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
		err = db.tasks.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating tasks collection in database %s: %v", db.conf.Database, err)
		}

		err = db.tasks.EnsureIndex(mgo.Index{
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
		err = db.nodes.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating nodes collection in database %s: %v", db.conf.Database, err)
		}

		err = db.nodes.EnsureIndex(mgo.Index{
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
func (db *MongoDB) Close() error {
	db.sess.Close()
	return nil
}
