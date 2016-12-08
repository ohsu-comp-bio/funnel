package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"tes/ga4gh"
	"tes/server"
)

func StartHttpProxy(rpcPort string, httpPort string, contentDir string) {
	//setup RESTful proxy
	grpcMux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	log.Println("Proxy connecting to localhost:" + rpcPort)
	err := ga4gh_task_exec.RegisterTaskServiceHandlerFromEndpoint(ctx, grpcMux, "localhost:" + rpcPort, opts)
	if err != nil {
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
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(contentDir, "index.html"))
		})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(contentDir))))

	r.PathPrefix("/v1/").Handler(grpcMux)
	log.Printf("Listening on port: %s\n", httpPort)
	http.ListenAndServe(":" + httpPort, r)
}

func main() {
	httpPort := flag.String("port", "8000", "HTTP Port")
	rpcPort := flag.String("rpc", "9090", "HTTP Port")
	storageDirArg := flag.String("storage", "storage", "Storage Dir")
	swiftArg := flag.Bool("swift", false, "Use SWIFT object store")
	taskDB := flag.String("db", "ga4gh_tasks.db", "Task DB File")

	flag.Parse()

	dir, _ := filepath.Abs(os.Args[0])
	contentDir := filepath.Join(dir, "..", "..", "share")

	//server meta-data
	storageDir, _ := filepath.Abs(*storageDirArg)
	var metaData = make(map[string]string)
	if !*swiftArg {
		metaData["storageType"] = "sharedFile"
		metaData["baseDir"] = storageDir
	} else {
		metaData["storageType"] = "swift"
	}

	//setup GRPC listener
	taski := tes_server.NewTaskBolt(*taskDB, metaData)

	server := tes_server.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	server.RegisterScheduleServer(taski)
	server.Start(*rpcPort)

	StartHttpProxy(*rpcPort, *httpPort, contentDir)
}
