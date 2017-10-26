package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
)

// DynamoDB provides handlers for gRPC endpoints
// Data is stored/retrieved from the Amazon DynamoDB NoSQL database.
type DynamoDB struct {
	client         *dynamodb.DynamoDB
	backend        compute.Backend
	partitionKey   string
	partitionValue string
	taskTable      string
	contentsTable  string
	stdoutTable    string
	stderrTable    string
}

// NewDynamoDB returns a new instance of DynamoDB, accessing the database at
// the given url, and including the given ServerConfig.
func NewDynamoDB(conf config.DynamoDB) (*DynamoDB, error) {
	awsConf := util.NewAWSConfigWithCreds(conf.Credentials.Key, conf.Credentials.Secret)
	awsConf.WithRegion(conf.Region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating dynamodb client: %v", err)
	}

	db := &DynamoDB{
		client:         dynamodb.New(sess),
		partitionKey:   "hid",
		partitionValue: "0",
		taskTable:      conf.TableBasename + "-task",
		contentsTable:  conf.TableBasename + "-contents",
		stdoutTable:    conf.TableBasename + "-stdout",
		stderrTable:    conf.TableBasename + "-stderr",
	}

	return db, nil
}

// Init creates tables in DynamoDB. If these tables already exist,
// a Debug level log is produced.
func (db *DynamoDB) Init(ctx context.Context) error {
	return db.createTables()
}

// WithComputeBackend configures the DynamoDB instance to use the given
// compute.Backend. The compute backend is responsible for dispatching tasks to
// schedulers / compute resources with its Submit method.
func (db *DynamoDB) WithComputeBackend(backend compute.Backend) {
	db.backend = backend
}
