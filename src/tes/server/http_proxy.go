package tes_server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"tes/ga4gh"
)

func HandleError(w http.ResponseWriter, req *http.Request, error string, code int) {
  fmt.Println(error)
  fmt.Println(req.URL)
  debug.PrintStack()
  http.Error(w, error, code)
}

func StartHttpProxy(rpcPort string, httpPort string, contentDir string) {
	//setup RESTful proxy
	grpcMux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	log.Println("HTTP proxy connecting to localhost:" + rpcPort)
	err := ga4gh_task_exec.RegisterTaskServiceHandlerFromEndpoint(ctx, grpcMux, "localhost:" + rpcPort, opts)
	if err != nil {
		fmt.Println("Register Error", err)

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
	log.Printf("HTTP API listening on port: %s\n", httpPort)
	http.ListenAndServe(":" + httpPort, r)
}
