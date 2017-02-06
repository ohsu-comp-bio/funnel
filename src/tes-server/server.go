package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"tes"
	"tes/scheduler"
	"tes/scheduler/condor"
	"tes/scheduler/dumblocal"
	"tes/scheduler/local"
	"tes/scheduler/openstack"
	"tes/server"
)

func main() {
	config := tes.DefaultConfig()

	var configArg string
	flag.StringVar(&configArg, "config", "", "Config File")
	flag.StringVar(&config.HTTPPort, "http-port", config.HTTPPort, "HTTP Port")
	flag.StringVar(&config.RPCPort, "rpc-port", config.RPCPort, "RPC Port")
	flag.StringVar(&config.DBPath, "db-path", config.DBPath, "Database path")
	flag.StringVar(&config.Scheduler, "scheduler", config.Scheduler, "Name of scheduler to enable")
	flag.Parse()

	tes.LoadConfigOrExit(configArg, &config)
	start(config)
}

func start(config tes.Config) {
	os.MkdirAll(config.WorkDir, 0755)
	// setup GRPC listener
	// TODO if another process has the db open, this will block and it is really
	//      confusing when you don't realize you have the db locked in another
	//      terminal somewhere. Would be good to timeout on startup here.
	taski := server.NewTaskBolt(config.DBPath, config.ServerConfig)

	s := server.NewGA4GHServer()
	s.RegisterTaskServer(taski)
	s.RegisterScheduleServer(taski)
	s.Start(config.RPCPort)

	// right now we only support a single scheduler at a time
	// enforced by tes.ValidateConfig()	above
	var sched scheduler.Scheduler
	switch strings.ToLower(config.Scheduler) {
	case "local":
		// TODO worker will stay alive if the parent process panics
		sched = local.NewScheduler(config)
	case "condor":
		sched = condor.NewScheduler(config)
	case "openstack":
		sched = openstack.NewScheduler(config)
	case "dumblocal":
		sched = dumblocal.NewScheduler(config)
	default:
		log.Printf("Error: unknown scheduler %s", config.Scheduler)
		return
	}
	go scheduler.StartScheduling(taski, sched)

	server.StartHTTPProxy(config.RPCPort, config.HTTPPort, config.ContentDir)
}
