package server

import (
	"context"
	"funnel/config"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"net/http"
	"runtime/debug"
)

func httpMux(ctx context.Context, conf config.Config) (*http.ServeMux, error) {

	// Set up HTTP proxy of gRPC API
	grpcMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	runtime.OtherErrorHandler = handleError

	var err error
	err = tes.RegisterTaskServiceHandlerFromEndpoint(
		ctx, grpcMux, conf.RPCAddress(), opts,
	)
	err = pbf.RegisterSchedulerHandlerFromEndpoint(
		ctx, grpcMux, conf.RPCAddress(), opts,
	)

	if err != nil {
		log.Error("Couldn't register services", err)
		return nil, err
	}

	// Serves dashboard and static files (html, img, css, etc.)
	fileServer := http.FileServer(http.Dir(conf.ContentDir))

	// Set up URL path handlers
	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("/v1/", grpcMux)

	// Set "cache-control: no-store" to disable response caching.
	// Without this, some servers (e.g. GCE) will cache a response from ListJobs, GetJob, etc.
	// which results in confusion about the stale data.
	if conf.DisableHTTPCache {
		mux.Handle("/v1/jobs", noCacheHandler(grpcMux))
	}
	return mux, nil
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
func noCacheHandler(next http.Handler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("cache-control", "no-store")
		next.ServeHTTP(resp, req)
	}
}
