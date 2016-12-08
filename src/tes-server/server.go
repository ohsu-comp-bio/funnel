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
	"tes"
	"tes/ga4gh"
	"tes/server"
	"tes/server/proto"
)

func main() {
	httpPort := flag.String("port", "8000", "HTTP Port")
	rpcPort := flag.String("rpc", "9090", "HTTP Port")
	storageDirArg := flag.String("storage", "", "Storage Dir")
	sharedDirArg := flag.String("shared", "", "Shared File System")
	swiftArg := flag.String("swift", "", "Use SWIFT object store")
	taskDB := flag.String("db", "ga4gh_tasks.db", "Task DB File")
	configFile := flag.String("config", "", "Config File")

	flag.Parse()

	dir, _ := filepath.Abs(os.Args[0])
	contentDir := filepath.Join(dir, "..", "..", "share")
	
	config := ga4gh_task_ref.ServerConfig{}
	if *configFile != "" {
		var err error
		config, err = tes.ParseConfigFile(*configFile)
		if err != nil {
			log.Println("Failure Reading Config")
			return
		}
	}
	if *storageDirArg != "" {
		//server meta-data
		storageDir, _ := filepath.Abs(*storageDirArg)
		fs := &ga4gh_task_ref.StorageConfig{Config:map[string]string{}}
		fs.Protocol = "fs"
		fs.Config["basedir"] = storageDir
		config.Storage = append(config.Storage, fs)
	}
	if *swiftArg != "" {
		fs := &ga4gh_task_ref.StorageConfig{Config:map[string]string{}}
		fs.Protocol = "swift"
		fs.Config["endpoint"] = *swiftArg
		config.Storage = append(config.Storage, fs)
	}
	if *sharedDirArg != "" {
		fs := &ga4gh_task_ref.StorageConfig{Config:map[string]string{}}
		fs.Protocol = "file"
		fs.Config["dirs"] = *sharedDirArg
		config.Storage = append(config.Storage, fs)
	}
	log.Printf("Config: %v\n", config)

	//setup GRPC listener
	taski := tes_server.NewTaskBolt(*taskDB, config)

	server := tes_server.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	server.RegisterScheduleServer(taski)
	server.Start(*rpcPort)

	//setup RESTful proxy
	grpcMux := runtime.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	log.Println("Proxy connecting to localhost:" + *rpcPort)
	err := ga4gh_task_exec.RegisterTaskServiceHandlerFromEndpoint(ctx, grpcMux, "localhost:"+*rpcPort, opts)
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
	log.Printf("Listening on port: %s\n", *httpPort)
	http.ListenAndServe(":"+*httpPort, r)
}
