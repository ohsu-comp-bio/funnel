package server

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/webdash"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"runtime/debug"
)

var log = logger.Sub("server")

// Server represents a Funnel server. The server handles
// RPC traffic via gRPC, HTTP traffic for the TES API,
// and also serves the web dashboard.
type Server struct {
	RPCAddress             string
	HTTPPort               string
	Password               string
	TaskServiceServer      tes.TaskServiceServer
	SchedulerServiceServer pbf.SchedulerServiceServer
	Handler                http.Handler
	DisableHTTPCache       bool
	DialOptions            []grpc.DialOption
}

// DefaultServer returns a new server instance.
func DefaultServer(conf config.Config) *Server {
	log.Debug("Server Config", "config.Config", conf)

	mux := http.NewServeMux()
	mux.Handle("/", webdash.Handler())

	return &Server{
		RPCAddress:       ":" + conf.RPCPort,
		HTTPPort:         conf.HTTPPort,
		Password:         conf.Server.Password,
		Handler:          mux,
		DisableHTTPCache: conf.DisableHTTPCache,
		DialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	}
}

// Serve starts the server and does not block. This will open TCP ports
// for both RPC and HTTP.
func (s *Server) Serve(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	// Open TCP connection for RPC
	lis, err := net.Listen("tcp", s.RPCAddress)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		// API auth check.
		grpc.UnaryInterceptor(newAuthInterceptor(s.Password)),
	)

	// Set up HTTP proxy of gRPC API
	mux := http.NewServeMux()
	grpcMux := runtime.NewServeMux()
	runtime.OtherErrorHandler = handleError

	// Set "cache-control: no-store" to disable response caching.
	// Without this, some servers (e.g. GCE) will cache a response from ListTasks, GetTask, etc.
	// which results in confusion about the stale data.
	// TODO BUG HERE, DisableHTTPCache is required.
	if s.DisableHTTPCache {
		mux.Handle("/v1/", disableCache(grpcMux))
	}

	if s.Handler != nil {
		mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
			s.Handler.ServeHTTP(resp, req)
		})
	}

	// Register TES service
	if s.TaskServiceServer != nil {
		log.Debug("Registering task service")
		tes.RegisterTaskServiceServer(grpcServer, s.TaskServiceServer)
		err := tes.RegisterTaskServiceHandlerFromEndpoint(
			ctx, grpcMux, s.RPCAddress, s.DialOptions,
		)
		if err != nil {
			return err
		}
	}

	// Register Scheduler RPC service
	if s.SchedulerServiceServer != nil {
		pbf.RegisterSchedulerServiceServer(grpcServer, s.SchedulerServiceServer)
		err := pbf.RegisterSchedulerServiceHandlerFromEndpoint(
			ctx, grpcMux, s.RPCAddress, s.DialOptions,
		)
		if err != nil {
			return err
		}
	}

	httpServer := &http.Server{
		Addr:    ":" + s.HTTPPort,
		Handler: mux,
	}

	var srverr error
	go func() {
		srverr = grpcServer.Serve(lis)
		cancel()
	}()

	go func() {
		srverr = httpServer.ListenAndServe()
		cancel()
	}()

	log.Info("Server listening",
		"httpPort", s.HTTPPort, "rpcAddress", s.RPCAddress,
	)

	<-ctx.Done()
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.TODO())

	return srverr
}

// handleError handles errors in the HTTP stack, logging errors, stack traces,
// and returning an HTTP error code.
func handleError(w http.ResponseWriter, req *http.Request, err string, code int) {
	log.Error("HTTP handler error", "error", err, "url", req.URL)
	debug.PrintStack()
	http.Error(w, err, code)
}

// Set a cache-control header that disables response caching
// and pass through to the next mux.
func disableCache(next http.Handler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(resp, req)
	}
}
