package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
)

// Client is a client for the scheduler gRPC service.
type Client interface {
	pbf.SchedulerServiceClient
	Close()
}

type client struct {
	pbf.SchedulerServiceClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler at a given address (e.g. "localhost:9090")
func NewClient(conf config.Worker) (Client, error) {
	log := log.WithFields("address", conf.ServerAddress)

	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	opts := util.DialOpts{}
	opts.Password(conf.ServerPassword)
	err := opts.TLS(conf.TLS.CertFile)

	if err != nil {
		log.Error("Client dial failed to setup TLS", err)
		return nil, err
	}

	conn, err := grpc.Dial(conf.ServerAddress, opts...)

	if err != nil {
		log.Error("Couldn't open RPC connection to scheduler", err)
		return nil, err
	}

	s := pbf.NewSchedulerServiceClient(conn)
	return &client{s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
