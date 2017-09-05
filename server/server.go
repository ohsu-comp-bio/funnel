// Package server handles serving HTTP/RPC APIs for TES, task logging, and scheduling,
// as well as database access.
package server

import (
	"context"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	tl "github.com/ohsu-comp-bio/funnel/proto/tasklogger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	webdash "github.com/ohsu-comp-bio/funnel/server/internal"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

var log = logger.Sub("server")

// Server represents a Funnel server. The server handles
// RPC traffic via gRPC, HTTP traffic for the TES API,
// and also serves the web dashboard.
type Server struct {
	Services struct {
		Tasks      tes.TaskServiceServer
		TaskLogger tl.TaskLoggerServiceServer
		Scheduler  pbs.SchedulerServiceServer
	}
	RPCAddress       string
	HTTPPort         string
	Password         string
	Handler          http.Handler
	DisableHTTPCache bool
	DialOptions      []grpc.DialOption
}

// DefaultServer returns a new server instance with defaults configured:
//
//   - RPC and HTTP ports are taken from the config.
//   - If present, conf.Password is used for HTTP basic auth.
//   - The RPC transport is insecure.
//   - The HTTP cache is disabled.
//   - The web dashboard is included.
//
// Services must be configured manually, e.g.
//   s := server.DefaultServer()
//   db, _ := server.NewTaskBolt(conf)
//   s.Services.Tasks = db
func DefaultServer(conf config.Server) *Server {
	log.Debug("Server Config", "config.Server", conf)

	mux := http.NewServeMux()
	mux.Handle("/", webdashHandler())

	return &Server{
		RPCAddress:       ":" + conf.RPCPort,
		HTTPPort:         conf.HTTPPort,
		Password:         conf.Password,
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
	if s.DisableHTTPCache {
		mux.Handle("/v1/", disableCache(grpcMux))
	}

	if s.Handler != nil {
		mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
			s.Handler.ServeHTTP(resp, req)
		})
	}

	// Register TES service
	if s.Services.Tasks != nil {
		tes.RegisterTaskServiceServer(grpcServer, s.Services.Tasks)
		err := tes.RegisterTaskServiceHandlerFromEndpoint(
			ctx, grpcMux, s.RPCAddress, s.DialOptions,
		)
		if err != nil {
			return err
		}
	}

	// Register TaskLogger service
	if s.Services.TaskLogger != nil {
		tl.RegisterTaskLoggerServiceServer(grpcServer, s.Services.TaskLogger)
	}

	// Register Scheduler RPC service
	if s.Services.Scheduler != nil {
		pbs.RegisterSchedulerServiceServer(grpcServer, s.Services.Scheduler)
		err := pbs.RegisterSchedulerServiceHandlerFromEndpoint(
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

// Handler handles static webdash files
func webdashHandler() *http.ServeMux {
	// Static files are bundled into webdash
	fs := http.FileServer(&assetfs.AssetFS{
		Asset:     webdash.Asset,
		AssetDir:  webdash.AssetDir,
		AssetInfo: webdash.AssetInfo,
		Prefix:    "webdash",
	})
	// Set up URL path handlers
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	return mux
}
