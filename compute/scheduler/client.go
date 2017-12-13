package scheduler

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
)

// Client is a client for the scheduler and event gRPC services.
type Client interface {
	events.EventServiceClient
	pbs.SchedulerServiceClient
	Close()
}

type client struct {
	events.EventServiceClient
	pbs.SchedulerServiceClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler and task logger services at a given address
// (e.g. "localhost:9090")
func NewClient(conf config.Server) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.RPCClientTimeout)
	defer cancel()

	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	// Create a connection for gRPC clients
	conn, err := grpc.DialContext(
		ctx,
		conf.RPCAddress(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		util.PerRPCPassword(conf.Password),
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't open RPC connection to the scheduler at %s: %s",
			conf.RPCAddress(), err)
	}

	e := events.NewEventServiceClient(conn)
	s := pbs.NewSchedulerServiceClient(conn)
	return &client{e, s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
