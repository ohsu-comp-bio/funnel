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
	"github.com/gengo/grpc-gateway/runtime"
	"ga4gh-tasks"
	"ga4gh-server"
	"runtime/debug"
	"log"
	"ga4gh-engine"
	"ga4gh-engine/scaling"
)



func main() {
	http_port := flag.String("port", "8000", "HTTP Port")
	rpc_port := flag.String("rpc", "9090", "HTTP Port")
	storage_dir_arg := flag.String("storage", "storage", "Storage Dir")
	task_db := flag.String("db", "ga4gh_tasks.db", "Task DB File")
	scaler_name := flag.String("scaler", "local", "Scaler")

	flag.Parse()
  
  	dir, _ := filepath.Abs(os.Args[0])
	content_dir := filepath.Join(dir, "..", "..", "share")

	config := map[string]string{}

	//get scaler
	scaler := ga4gh_engine_scaling.ScalingMethods[*scaler_name](config)

	//server meta-data
	storage_dir, _ := filepath.Abs(*storage_dir_arg)
	meta_data := map[string]string{ "storageType" : "sharedFile", "baseDir" : storage_dir }

	//setup GRPC listener
	taski := ga4gh_task.NewTaskBolt(*task_db, meta_data) //ga4gh_task.NewTaskImpl()

	//setup scheduler
	scheduler := ga4gh_taskengine.Scheduler(taski, scaler)

	server := ga4gh_task.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	server.RegisterScheduleServer(scheduler)
	server.Start(*rpc_port)

	//setup RESTful proxy
	grpc_mux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	log.Println("Proxy connecting to localhost:" + *rpc_port )
	err := ga4gh_task_exec.RegisterTaskServiceHandlerFromEndpoint(ctx, grpc_mux, "localhost:" + *rpc_port, opts)
	if (err != nil) {
		fmt.Println("Register Error", err)

	}
	r := mux.NewRouter()

	runtime.OtherErrorHandler = func(w http.ResponseWriter, req *http.Request, error string, code int) {
		fmt.Println(error)
		fmt.Println(req.URL)
		debug.PrintStack()
		http.Error(w, error, code)
	}
	// Routes consist of a path and a handler function
	r.HandleFunc("/",
		func (w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(content_dir, "index.html"))
		})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(content_dir))))

	r.PathPrefix("/v1/").Handler(grpc_mux)
	log.Printf("Listening on port: %s\n", *http_port)
	http.ListenAndServe(":" + *http_port, r)
}