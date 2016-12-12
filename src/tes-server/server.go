package main

import (
	"flag"
	"os"
	"path/filepath"
	"tes/server"
)

func main() {
	httpPort := flag.String("port", "8000", "HTTP Port")
	rpcPort := flag.String("rpc", "9090", "TCP+RPC Port")
	storageDirArg := flag.String("storage", "storage", "Storage Dir")
	swiftArg := flag.Bool("swift", false, "Use SWIFT object store")
	taskDB := flag.String("db", "ga4gh_tasks.db", "Task DB File")

	flag.Parse()

	// server meta-data
	storageDir, _ := filepath.Abs(*storageDirArg)
	var metaData = make(map[string]string)
	if !*swiftArg {
		metaData["storageType"] = "sharedFile"
		metaData["baseDir"] = storageDir
	} else {
		metaData["storageType"] = "swift"
	}

	// setup GRPC listener
	taski := tes_server.NewTaskBolt(*taskDB, metaData)

	server := tes_server.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	server.RegisterScheduleServer(taski)
	server.Start(*rpcPort)

	// Path to HTML content
	dir, _ := filepath.Abs(os.Args[0])
	contentDir := filepath.Join(dir, "..", "..", "share")

	tes_server.StartHttpProxy(*rpcPort, *httpPort, contentDir)
}
