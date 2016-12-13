package main

import (
	"flag"
	"os"
	"path/filepath"
	//"runtime/debug"
	"tes"
	//"tes/ga4gh"
	"log"
	"tes/server"
	"tes/server/proto"
)

func main() {
	httpPort := flag.String("port", "8000", "HTTP Port")
	rpcPort := flag.String("rpc", "9090", "TCP+RPC Port")
	storageDirArg := flag.String("storage", "storage", "Storage Dir")
	sharedDirArg := flag.String("shared", "", "Shared File System")
	s3Arg := flag.String("s3", "", "Use S3 object store")
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
		fs := &ga4gh_task_ref.StorageConfig{Config: map[string]string{}}
		fs.Protocol = "fs"
		fs.Config["basedir"] = storageDir
		config.Storage = append(config.Storage, fs)
	}
	if *s3Arg != "" {
		fs := &ga4gh_task_ref.StorageConfig{Config: map[string]string{}}
		fs.Protocol = "s3"
		fs.Config["endpoint"] = *s3Arg
		config.Storage = append(config.Storage, fs)
	}
	if *sharedDirArg != "" {
		fs := &ga4gh_task_ref.StorageConfig{Config: map[string]string{}}
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

	tes_server.StartHttpProxy(*rpcPort, *httpPort, contentDir)
}
