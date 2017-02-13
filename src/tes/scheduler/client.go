package scheduler

import (
	"google.golang.org/grpc"
	"tes/config"
	pbr "tes/server/proto"
)

// Client is a client for the scheduler gRPC service.
type Client struct {
	pbr.SchedulerClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler at a given address (e.g. "localhost:9090")
func NewClient(conf config.Worker) (*Client, error) {
	conn, err := NewRPCConnection(conf.ServerAddress)
	if err != nil {
		log.Error("Couldn't connect to schduler", err)
		return nil, err
	}

	s := pbr.NewSchedulerClient(conn)
	return &Client{s, conn}, nil
}

// Close closes the client connection.
func (client *Client) Close() {
	client.conn.Close()
}
