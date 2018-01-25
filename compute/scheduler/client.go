package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	util "github.com/ohsu-comp-bio/funnel/util/rpc"
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
func NewClient(ctx context.Context, conf config.Server) (Client, error) {
	conn, err := util.Dial(ctx, conf)
	if err != nil {
		return nil, err
	}
	e := events.NewEventServiceClient(conn)
	s := pbs.NewSchedulerServiceClient(conn)
	return &client{e, s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
