package tes_server

import (
	"google.golang.org/grpc"
	"log"
)

func NewRpcConnection(address string) (*grpc.ClientConn, error) {
	// TODO write a test for the case when the scheduler service goes
	//      down for 5 seconds (time > scaler poll time). Does this
	//      client handle that connection interruption ok?
	// TODO if this can't connect initially, should it retry?
	// Create a connection for gRPC clients
	serverAddr := "localhost:9090"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Printf("Can't open RPC connection to %s", address)
		log.Println(err.Error())
		return nil, err
		// TODO give up after max retries (configurable)
	}
	return conn, nil
}
