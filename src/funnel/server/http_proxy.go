package server

import (
	"funnel/proto/tes"
	pbf "funnel/proto/funnel"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/http"
	"path/filepath"
	"runtime/debug"
)

// HandleError handles errors in the HTTP stack, logging errors, stack traces,
// and returning an HTTP error code.
func HandleError(w http.ResponseWriter, req *http.Request, err string, code int) {
	log.Error("HTTP handler error", "error", err, "url", req.URL)
	debug.PrintStack()
	http.Error(w, err, code)
}

// StartHTTPProxy starts the HTTP proxy. It listens requests on the given HTTP port,
// and proxies the requests off to the given RPC port. The contentDir defines the
// location of web dashboard static files.
func StartHTTPProxy(serverAddress string, httpPort string, contentDir string) {
	//setup RESTful proxy
	grpcMux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	url := serverAddress
	log.Info("HTTP proxy listening gRPC", "url", url)
	err := tes.RegisterTaskServiceHandlerFromEndpoint(ctx, grpcMux, url, opts)
	if err != nil {
		log.Error("Couldn't register Task Service", "error", err)
	}
	serr := pbf.RegisterSchedulerHandlerFromEndpoint(ctx, grpcMux, url, opts)
	if serr != nil {
		log.Error("Couldn't register Scheduler service HTTP proxy", "error", err)
	}
	r := mux.NewRouter()

	runtime.OtherErrorHandler = HandleError
	// Routes consist of a path and a handler function
	r.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(contentDir, "index.html"))
		})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(contentDir))))

	r.PathPrefix("/v1/").Handler(grpcMux)
	log.Info("HTTP API listening", "port", httpPort)
	http.ListenAndServe(":"+httpPort, r)
}
