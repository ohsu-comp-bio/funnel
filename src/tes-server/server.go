package main

import (
	"flag"
	"os"
	"strings"
	"tes"
	"tes/logger"
	"tes/scheduler"
	"tes/scheduler/condor"
	"tes/scheduler/dumblocal"
	"tes/scheduler/local"
	"tes/scheduler/openstack"
	"tes/server"
)

var log = logger.New("tes-server")

func main() {
	config := tes.DefaultConfig()

	var configArg string
	flag.StringVar(&configArg, "config", "", "Config File")
	flag.StringVar(&config.HTTPPort, "http-port", config.HTTPPort, "HTTP Port")
	flag.StringVar(&config.RPCPort, "rpc-port", config.RPCPort, "RPC Port")
	flag.StringVar(&config.DBPath, "db-path", config.DBPath, "Database path")
	flag.StringVar(&config.Scheduler, "scheduler", config.Scheduler, "Name of scheduler to enable")
	flag.StringVar(&config.LogLevel, "logging", config.LogLevel, "Level of logging")
	flag.Parse()

	tes.LoadConfigOrExit(configArg, &config)
	logger.SetLogLevel(config.LogLevel)
	start(config)
}

func start(config tes.Config) {
	os.MkdirAll(config.WorkDir, 0755)

	taski, err := server.NewTaskBolt(config.DBPath, config.ServerConfig)
	if err != nil {
		log.Error("Couldn't open database", err)
		return
	}

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
		log.Error("Unknown scheduler",
			"scheduler", config.Scheduler)
		return
	}
	go scheduler.StartScheduling(taski, sched)

	server.StartHTTPProxy(config.RPCPort, config.HTTPPort, config.ContentDir)
}
