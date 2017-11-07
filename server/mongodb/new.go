package mongodb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"golang.org/x/net/context"
	mgo "gopkg.in/mgo.v2"
)

// MongoDB provides an MongoDB database server backend.
type MongoDB struct {
	sess    *mgo.Session
	backend compute.Backend
	conf    config.MongoDB
	tasks   *mgo.Collection
	// nodes   *mgo.Collection
	// events  *mgo.Collection
}

// NewMongoDB returns a new MongoDB instance.
func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	sess, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    conf.Addrs,
		Username: conf.Username,
		Password: conf.Password,
		Database: conf.Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		return nil, err
	}
	return &MongoDB{
		sess:  sess,
		conf:  conf,
		tasks: sess.DB(conf.Database).C("tasks"),
	}, nil
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init(ctx context.Context) error {
	names, err := db.sess.DB(db.conf.Database).CollectionNames()
	if err != nil {
		return fmt.Errorf("error listing collection names in database %s: %v", db.conf.Database, err)
	}
	for _, n := range names {
		if n == "tasks" {
			return nil
		}
	}
	err = db.tasks.Create(&mgo.CollectionInfo{})
	if err != nil {
		return fmt.Errorf("error creating collection in database %s: %v", db.conf.Database, err)
	}
	return db.tasks.EnsureIndex(mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	})
}

// WithComputeBackend configures the MongoDB instance to use the given
// compute.Backend. The compute backend is responsible for dispatching tasks to
// schedulers / compute resources with its Submit method.
func (db *MongoDB) WithComputeBackend(backend compute.Backend) {
	db.backend = backend
}
