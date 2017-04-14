package server

import (
	"context"
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

var log = logger.New("server")

// Server represents a Funnel server. The server handles
// RPC traffic via gRPC, HTTP traffic for the TES API,
// and also serves the web dashboard.
type Server struct {
	conf       config.Config
	grpcServer *grpc.Server
	httpServer *http.Server
}

// NewServer returns a new Server instance.
func NewServer(db Database, conf config.Config) (*Server, error) {
	log.Debug("Server Config", "config.Config", conf)

	grpcServer := grpc.NewServer()
	tes.RegisterTaskServiceServer(grpcServer, db)
	pbf.RegisterSchedulerServer(grpcServer, db)

	httpServer := &http.Server{
		Addr: ":" + conf.HTTPPort,
	}

	return &Server{conf, grpcServer, httpServer}, nil
}

// Start starts the server and does not block. This will open TCP ports
// for both RPC and HTTP.
func (s *Server) Start(ctx context.Context) error {

	lis, err := net.Listen("tcp", ":"+s.conf.RPCPort)
	if err != nil {
		return err
	}

	httpHandler, err := httpMux(ctx, s.conf)
	if err != nil {
		return err
	}
	s.httpServer.Handler = httpHandler

	log.Info("RPC server listening", "port", s.conf.RPCPort)
	go func() {
		err := s.grpcServer.Serve(lis)
		log.Error("RPC server error", err)
	}()

	log.Info("HTTP server listening",
		"port", s.conf.HTTPPort, "rpcAddress", s.conf.RPCAddress(),
	)
	// TODO how do we handle errors returned from grpcServer.Serve()
	//      httpServer.ListenAndServe()
	go func() {
		err := s.httpServer.ListenAndServe()
		log.Error("HTTP server error", err)
	}()
	return nil
}

// Stop stops the server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
	s.httpServer.Shutdown(context.TODO())
}
