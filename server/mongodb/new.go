package mongodb

import (
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"golang.org/x/net/context"
	mgo "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	// "time"
)

type MongoDB struct {
	client  *mgo.Session
	backend compute.Backend
	conf    config.MongoDB
}

func NewMongoDB(conf config.MongoDB) (*MongoDB, error) {
	sess, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs: conf.Addrs,
		// Username: conf.Username,
		// Password: conf.Password,
		Database: conf.Database,
		// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
		// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
		// },
	})
	if err != nil {
		return nil, err
	}
	return &MongoDB{
		client: sess,
		conf:   conf,
	}, nil
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init(ctx context.Context) error {
	return db.client.DB(db.conf.Database).C(db.conf.Collection).Create(&mgo.CollectionInfo{})
}

// WithComputeBackend configures the MongoDB instance to use the given
// compute.Backend. The compute backend is responsible for dispatching tasks to
// schedulers / compute resources with its Submit method.
func (db *MongoDB) WithComputeBackend(backend compute.Backend) {
	db.backend = backend
}
