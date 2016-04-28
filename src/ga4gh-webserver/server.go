package main

import (
	"os"
	"fmt"
	"flag"
	"net/http"
	"path/filepath"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"ga4gh-tasks"
	"ga4gh-server"
)


func main() {
	http_port := flag.String("port", "8000", "HTTP Port")
	rpc_port := flag.String("rpc", "9090", "HTTP Port")
	flag.Parse()
  
  	dir, _ := filepath.Abs(os.Args[0])
	content_dir := filepath.Join(dir, "..", "..", "share")

	//setup GRPC listener

	taski := ga4gh_task.NewTaskImpl()

	server := ga4gh_task.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	go server.Run(rpc_port)

	fmt.Printf("Listening on port: %s\n", *http_port)
	
	r := mux.NewRouter()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{}
	_ := ga4gh_task_exec.RegisterTaskServiceHandlerFromEndpoint(ctx, r, *server, opts)

	// Routes consist of a path and a handler function.
	r.HandleFunc("/",
		func (w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(content_dir, "index.html"))
		})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	http.ListenAndServe(":" + *http_port, r)
}