package main

import (
	"flag"
	"github.com/imdario/mergo"
	"os"
	"strings"
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	"tes/scheduler/condor"
	"tes/scheduler/gce"
	"tes/scheduler/local"
	"tes/server"
)

var log = logger.New("tes-server")

func main() {
	var cliconf config.Config
	var confArg string
	flag.StringVar(&confArg, "config", "", "Config File")
	flag.StringVar(&cliconf.HostName, "hostname", cliconf.HostName, "Host name or IP")
	flag.StringVar(&cliconf.HTTPPort, "http-port", cliconf.HTTPPort, "HTTP Port")
	flag.StringVar(&cliconf.RPCPort, "rpc-port", cliconf.RPCPort, "RPC Port")
	flag.StringVar(&cliconf.DBPath, "db-path", cliconf.DBPath, "Database path")
	flag.StringVar(&cliconf.Scheduler, "scheduler", cliconf.Scheduler, "Name of scheduler to enable")
	flag.StringVar(&cliconf.LogLevel, "logging", cliconf.LogLevel, "Level of logging")
	flag.Parse()

	conf := config.DefaultConfig()
	config.LoadConfigOrExit(confArg, &conf)

	// file vals <- cli val
	err := mergo.MergeWithOverwrite(&conf, cliconf)
	if err != nil {
		panic(err)
	}

	// make sure the proper defaults are set
	conf.Worker = config.WorkerInheritConfigVals(conf)

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
	case "condor":
		sched = condor.NewScheduler(conf)
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

	// If the scheduler implements the Scaler interface,
	// start a scaler loop
	if s, ok := sched.(scheduler.Scaler); ok {
		go scheduler.ScaleLoop(taski, s, conf)
	}

	// TODO if port 8000 is already busy, does this lock up silently?
	server.StartHTTPProxy(conf.HostName+":"+conf.RPCPort, conf.HTTPPort, conf.ContentDir)
}
