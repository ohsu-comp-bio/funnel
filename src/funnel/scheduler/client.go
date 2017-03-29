package scheduler

import (
	"funnel/config"
	pbf "funnel/proto/funnel"
	"google.golang.org/grpc"
)

// Client is a client for the scheduler gRPC service.
type Client interface {
	pbf.SchedulerClient
	Close()
}

type client struct {
	pbf.SchedulerClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler at a given address (e.g. "localhost:9090")
func NewClient(conf config.Worker) (Client, error) {
	conn, err := NewRPCConnection(conf.ServerAddress)
	if err != nil {
		log.Error("Couldn't connect to scheduler", err)
		return nil, err
	}

	s := pbf.NewSchedulerClient(conn)
	return &client{s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}

// NewRPCConnection returns a gRPC ClientConn, or an error.
// Use this for getting a connection for gRPC clients.
func NewRPCConnection(address string) (*grpc.ClientConn, error) {
	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	// Create a connection for gRPC clients
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Error("Couldn't open RPC connection",
			"error", err,
			"address", address,
		)
		return nil, err
	}
	return conn, nil
}
