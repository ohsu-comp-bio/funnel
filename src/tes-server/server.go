package main

import (
	"flag"
	"os"
	"strings"
	"tes/config"
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
	conf := config.DefaultConfig()

	var confArg string
	flag.StringVar(&confArg, "config", "", "Config File")
	flag.StringVar(&conf.HTTPPort, "http-port", conf.HTTPPort, "HTTP Port")
	flag.StringVar(&conf.RPCPort, "rpc-port", conf.RPCPort, "RPC Port")
	flag.StringVar(&conf.DBPath, "db-path", conf.DBPath, "Database path")
	flag.StringVar(&conf.Scheduler, "scheduler", conf.Scheduler, "Name of scheduler to enable")
	flag.StringVar(&conf.LogLevel, "logging", conf.LogLevel, "Level of logging")
	flag.Parse()

	config.LoadConfigOrExit(confArg, &conf)
	logger.SetLevel(conf.LogLevel)
	start(conf)
}

func start(conf config.Config) {
	os.MkdirAll(conf.WorkDir, 0755)

	taski, err := server.NewTaskBolt(conf)
	if err != nil {
		log.Error("Couldn't open database", err)
		return
	}

	s := server.NewGA4GHServer()
	s.RegisterTaskServer(taski)
	s.RegisterScheduleServer(taski)
	s.Start(conf.RPCPort)

	var sched scheduler.Scheduler
	switch strings.ToLower(conf.Scheduler) {
	case "local":
		// TODO worker will stay alive if the parent process panics
		sched = local.NewScheduler(conf)
	case "condor":
		sched = condor.NewScheduler(conf)
	case "openstack":
		sched = openstack.NewScheduler(conf)
	case "dumblocal":
		sched = dumblocal.NewScheduler(conf)
	default:
		log.Error("Unknown scheduler",
			"scheduler", conf.Scheduler)
		return
	}
	go scheduler.StartScheduling(taski, sched, conf.Worker.NewJobPollRate)

	server.StartHTTPProxy(conf.RPCPort, conf.HTTPPort, conf.ContentDir)
}
