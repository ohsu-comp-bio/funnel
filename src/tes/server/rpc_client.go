package tes_server

import (
	"google.golang.org/grpc"
	"log"
)

// NewRpcConnection returns a gRPC ClientConn, or an error.
// Use this for getting a connection for gRPC clients.
func NewRpcConnection(address string) (*grpc.ClientConn, error) {
	// TODO if this can't connect initially, should it retry?
	//      give up after max retries? Does grpc.Dial already do this?
	// Create a connection for gRPC clients
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Printf("Can't open RPC connection to %s", address)
		log.Println(err.Error())
		return nil, err
	}
	return conn, nil
}
