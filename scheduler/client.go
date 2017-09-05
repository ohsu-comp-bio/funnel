package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	tl "github.com/ohsu-comp-bio/funnel/proto/tasklogger"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
)

// Client is a client for the scheduler and task logger gRPC services.
type Client interface {
	tl.TaskLoggerServiceClient
	pbs.SchedulerServiceClient
	Close()
}

type client struct {
	tl.TaskLoggerServiceClient
	pbs.SchedulerServiceClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler and task logger services at a given address
// (e.g. "localhost:9090")
func NewClient(conf config.Scheduler) (Client, error) {
	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	// Create a connection for gRPC clients
	conn, err := grpc.Dial(conf.Node.ServerAddress,
		grpc.WithInsecure(),
		util.PerRPCPassword(conf.Node.ServerPassword),
	)

	if err != nil {
		log.Error("Couldn't open RPC connection to the scheduler",
			"error", err,
			"address", conf.Node.ServerAddress,
		)
		return nil, err
	}

	t := tl.NewTaskLoggerServiceClient(conn)
	s := pbs.NewSchedulerServiceClient(conn)
	return &client{t, s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
