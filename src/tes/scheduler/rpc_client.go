package scheduler

import (
	"google.golang.org/grpc"
)

// NewRPCConnection returns a gRPC ClientConn, or an error.
// Use this for getting a connection for gRPC clients.
func NewRPCConnection(address string) (*grpc.ClientConn, error) {
	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	// Create a connection for gRPC clients
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
    log.Error("Couldn't open RPC connection",
      "error",   err,
      "address", address,
    )
		return nil, err
	}
	return conn, nil
}
