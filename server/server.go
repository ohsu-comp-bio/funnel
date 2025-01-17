// Package server contains code for serving the Funnel API, and accessing database backends.
package server

import (
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/golang/gddo/httputil"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/webdash"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server represents a Funnel server. The server handles
// RPC traffic via gRPC, HTTP traffic for the TES API,
// and also serves the web dashboard.
type Server struct {
	RPCAddress       string
	HTTPPort         string
	BasicAuth        []config.BasicCredential
	OidcAuth         config.OidcAuth
	Tasks            tes.TaskServiceServer
	Events           events.EventServiceServer
	Nodes            scheduler.SchedulerServiceServer
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

// customErrorHandler is a custom error handler for the gRPC gateway
// Returns '400' for invalid backend parameters and '500' for all other errors
// Required for TES Compliance Tests
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"error": "failed to process the request"}`

	st, ok := status.FromError(err)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fallback))
		return
	}

	// Map specific gRPC error codes to HTTP status codes
	switch st.Code() {
	case codes.Unauthenticated:
		w.WriteHeader(http.StatusUnauthorized) // 401
	case codes.PermissionDenied:
		w.WriteHeader(http.StatusForbidden) // 403
	case codes.NotFound:
		// Special case for missing tasks (TES Compliance Suite)
		if strings.Contains(st.Message(), "task not found") {
			w.WriteHeader(http.StatusInternalServerError) // 500
		} else {
			w.WriteHeader(http.StatusNotFound) // 404
		}
	default:
		if strings.Contains(st.Message(), "backend parameters not supported") {
			w.WriteHeader(http.StatusBadRequest) // 400
		} else {
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
	}

	// Write the error message
	jErr := JSONError{Error: st.Message()}
	jErrBytes, mErr := marshaler.Marshal(jErr)
	if mErr != nil {
		w.Write([]byte(fallback))
		return
	}
	w.Write(jErrBytes)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type JSONError struct {
	Error string `json:"error"`
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

	auth := NewAuthentication(s.BasicAuth, s.OidcAuth)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// API auth check.
				auth.Interceptor,
				newDebugInterceptor(s.Log),
			),
		),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Set up HTTP proxy of gRPC API
	mux := http.NewServeMux()

	marsh := NewMarshaler()
	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marsh), runtime.WithErrorHandler(customErrorHandler))

	// m := protojson.MarshalOptions{
	// 	Indent:          "  ",
	// 	EmitUnpopulated: true,
	// 	UseProtoNames:   true,
	// }
	// u := protojson.UnmarshalOptions{}
	// grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
	// 	MarshalOptions:   m,
	// 	UnmarshalOptions: u,
	// }))

	//runtime.OtherErrorHandler = s.handleError //TODO: Review effects

	dashmux := http.NewServeMux()
	dashmux.Handle("/", webdash.RootHandler())
	dashfs := webdash.FileServer()
	mux.Handle("/favicon.ico", dashfs)
	mux.Handle("/manifest.json", dashfs)
	mux.Handle("/health.html", dashfs)
	mux.HandleFunc("/healthz", healthHandler)
	mux.Handle("/static/", dashfs)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/login/token", auth.EchoTokenHandler)

	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// TODO this doesnt handle all routes
		if s.OidcAuth.ServiceConfigURL == "" && len(s.BasicAuth) > 0 {
			resp.Header().Set("WWW-Authenticate", "Basic")
		}
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
		scheduler.RegisterSchedulerServiceServer(grpcServer, s.Nodes)
		err := scheduler.RegisterSchedulerServiceHandlerFromEndpoint(
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
