package main

import (
	"flag"
	"os"
	"strings"
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	//"tes/scheduler/condor"
	"tes/scheduler/gce"
	"tes/scheduler/local"
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

	taski, dberr := server.NewTaskBolt(conf)
	if dberr != nil {
		log.Error("Couldn't open database", dberr)
		return
	}

	s := server.NewGA4GHServer()
	s.RegisterTaskServer(taski)
	s.RegisterScheduleServer(taski)
	go s.Start(conf.RPCPort)

	var sched scheduler.Scheduler
	var err error
	switch strings.ToLower(conf.Scheduler) {
	case "local":
		// TODO worker will stay alive if the parent process panics
		sched, err = local.NewScheduler(conf)
	//case "condor":
	//sched = condor.NewScheduler(conf)
	case "gce":
		sched, err = gce.NewScheduler(conf)
	//case "openstack":
	//sched = openstack.NewScheduler(conf)
	default:
		log.Error("Unknown scheduler", "scheduler", conf.Scheduler)
		return
	}

	if err != nil {
		log.Error("Couldn't create scheduler", err)
		return
	}

	go scheduler.ScheduleLoop(taski, sched, conf)

	// TODO if port 8000 is already busy, does this lock up silently?
	server.StartHTTPProxy(conf.RPCPort, conf.HTTPPort, conf.ContentDir)
}
