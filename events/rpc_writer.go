package events

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
)

// RPCWriter is a type which writes Events to RPC.
type RPCWriter struct {
	client EventServiceClient
}

// NewRPCWriter returns a new RPCWriter instance.
func NewRPCWriter(conf config.Server) (*RPCWriter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.RPCClientTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		conf.RPCAddress(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		util.PerRPCPassword(conf.Password),
	)
	if err != nil {
		return nil, err
	}
	cli := NewEventServiceClient(conn)

	return &RPCWriter{cli}, nil
}

// WriteEvent writes the event to the server via gRPC.
// The RPC call may timeout, based on the timeout given by the configuration in NewRPCWriter.
func (r *RPCWriter) WriteEvent(ctx context.Context, e *Event) error {
	_, err := r.client.WriteEvent(ctx, e)
	return err
}
