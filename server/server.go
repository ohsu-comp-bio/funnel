// Package server contains code for serving the Funnel API, and accessing database backends.
package server

import (
	"net"
	"net/http"

	"github.com/golang/gddo/httputil"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/ohsu-comp-bio/funnel/compute/builtin"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/webdash"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server represents a Funnel server. The server handles
// RPC traffic via gRPC, HTTP traffic for the TES API,
// and also serves the web dashboard.
type Server struct {
	RPCAddress       string
	HTTPPort         string
	User             string
	Password         string
	Tasks            tes.TaskServiceServer
	Events           events.EventServiceServer
	Nodes            builtin.SchedulerServiceServer
	DisableHTTPCache bool
	Log              *logger.Logger
}

// Return a new interceptor function that logs all requests at the Debug level
func newDebugInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	// Return a function that is the interceptor.
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		log.Debug(
			"received: "+info.FullMethod,
			"request", req,
		)
		resp, err := handler(ctx, req)
		log.Debug(
			"responding: "+info.FullMethod,
			"resp", resp,
			"err", err,
		)
		return resp, err
	}
}

// Run runs the server, listening on two TCP ports for HTTP and RPC traffic.
// Run blocks until the context is canceled, or an error occurs.
func (s *Server) Run(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	// Open TCP connection for RPC
	lis, err := net.Listen("tcp", s.RPCAddress)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// API auth check.
				newAuthInterceptor(s.User, s.Password),
				newDebugInterceptor(s.Log),
			),
		),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// Set up HTTP proxy of gRPC API
	mux := http.NewServeMux()
	mar := runtime.JSONPb(tes.Marshaler)
	grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption("*/*", &mar))
	runtime.OtherErrorHandler = s.handleError

	dashmux := http.NewServeMux()
	dashmux.Handle("/", webdash.RootHandler())
	dashfs := webdash.FileServer()
	mux.Handle("/favicon.ico", dashfs)
	mux.Handle("/static/", http.StripPrefix("/static/", dashfs))

	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {

		switch negotiate(req) {
		case "html":
			// HTML was requested (by the browser)
			dashmux.ServeHTTP(resp, req)
		default:
			// Set "cache-control: no-store" to disable response caching.
			// Without this, some servers (e.g. GCE) will cache a response from ListTasks, GetTask, etc.
			// which results in confusion about the stale data.
			if s.DisableHTTPCache {
				resp.Header().Set("Cache-Control", "no-store")
			}
			grpcMux.ServeHTTP(resp, req)
		}
	})

	// Register TES service
	if s.Tasks != nil {
		tes.RegisterTaskServiceServer(grpcServer, s.Tasks)
		err := tes.RegisterTaskServiceHandlerFromEndpoint(
			ctx, grpcMux, s.RPCAddress, dialOpts,
		)
		if err != nil {
			return err
		}
	}

	// Register Events service
	if s.Events != nil {
		events.RegisterEventServiceServer(grpcServer, s.Events)
	}

	// Register Scheduler RPC service
	if s.Nodes != nil {
		builtin.RegisterSchedulerServiceServer(grpcServer, s.Nodes)
		err := builtin.RegisterSchedulerServiceHandlerFromEndpoint(
			ctx, grpcMux, s.RPCAddress, dialOpts,
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

	s.Log.Info("Server listening",
		"httpPort", s.HTTPPort, "rpcAddress", s.RPCAddress,
	)

	<-ctx.Done()
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.TODO())

	return srverr
}

// handleError handles errors in the HTTP stack, logging errors, stack traces,
// and returning an HTTP error code.
func (s *Server) handleError(w http.ResponseWriter, req *http.Request, err string, code int) {
	s.Log.Error("HTTP handler error", "error", err, "url", req.URL)
	http.Error(w, err, code)
}

// negotiate determines the response type based on request headers and parameters.
// Returns either "html" or "json".
func negotiate(req *http.Request) string {
	// Allow overriding the type from a URL parameter.
	// /v1/tasks?json will force a JSON response.
	q := req.URL.Query()
	if _, html := q["html"]; html {
		return "html"
	}
	if _, json := q["json"]; json {
		return "json"
	}
	// Content negotiation means that both the dashboard's HTML and the API's JSON
	// may be served at the same path.
	// In Go 1.10 we'll be able to move to a core library for this,
	// https://github.com/golang/go/issues/19307
	switch httputil.NegotiateContentType(req, []string{"text/*", "text/html"}, "text/*") {
	case "text/html":
		return "html"
	default:
		return "json"
	}
}
