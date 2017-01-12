package main

import (
	"flag"
	"os"
	"path/filepath"
	//"runtime/debug"
	"tes"
	//"tes/ga4gh"
	"log"
	"tes/scheduler"
	"tes/scheduler/condor"
	"tes/scheduler/dumblocal"
	"tes/scheduler/local"
	"tes/scheduler/openstack"
	"tes/server"
)

func main() {
	httpPort := flag.String("port", "8000", "HTTP Port")
	rpcPort := flag.String("rpc", "9090", "TCP+RPC Port")
	taskDB := flag.String("db", "ga4gh_tasks.db", "Task DB File")
	schedArg := flag.String("sched", "local", "Scheduler")
	configFile := flag.String("config", "", "Config File")

	flag.Parse()

	dir, _ := filepath.Abs(os.Args[0])
	contentDir := filepath.Join(dir, "..", "..", "share")

	config := tes.Config{}
	tes.LoadConfigOrExit(*configFile, &config)

	//setup GRPC listener
	// TODO if another process has the db open, this will block and it is really
	//      confusing when you don't realize you have the db locked in another
	//      terminal somewhere. Would be good to timeout on startup here.
	taski := tes_server.NewTaskBolt(*taskDB, config.ServerConfig)

	server := tes_server.NewGA4GHServer()
	server.RegisterTaskServer(taski)
	server.RegisterScheduleServer(taski)
	server.Start(*rpcPort)

	var sched scheduler.Scheduler
	switch *schedArg {
	case "local":
		// TODO worker will stay alive if the parent process panics
		sched = local.NewScheduler(config)
	case "condor":
		sched = condor.NewScheduler(config)
	case "openstack":
		sched = openstack.NewScheduler(config)
	case "dumblocal":
		sched = dumblocal.NewScheduler(4)
	default:
		log.Printf("Error: unknown scheduler %s", *schedArg)
		return
	}
	go scheduler.StartScheduling(taski, sched)

	tes_server.StartHttpProxy(*rpcPort, *httpPort, contentDir)
}
