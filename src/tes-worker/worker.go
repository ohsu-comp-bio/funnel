package main

import (
	"flag"
	"github.com/imdario/mergo"
	"os"
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	"tes/worker"
)

var log = logger.New("tes-worker")

func main() {
	var cliconf config.Worker
	var confArg string
	flag.StringVar(&confArg, "config", "", "Config File")
	flag.StringVar(&cliconf.ID, "id", cliconf.ID, "Worker ID")
	flag.StringVar(&cliconf.ServerAddress, "server-address", cliconf.ServerAddress, "Server address")
	flag.StringVar(&cliconf.WorkDir, "work-dir", cliconf.WorkDir, "Working Directory")
	flag.DurationVar(&cliconf.Timeout, "timeout", cliconf.Timeout, "Timeout in seconds")
	flag.StringVar(&cliconf.LogPath, "log-path", cliconf.LogPath, "File path to write logs to")
	flag.Parse()

	dconf := config.DefaultConfig()
	conf := config.WorkerDefaultConfig(dconf)
	config.LoadConfigOrExit(confArg, &conf)

	// file vals <- cli val
	err := mergo.MergeWithOverwrite(&conf, cliconf)
	if err != nil {
		panic(err)
	}

	start(conf)
}

func start(conf config.Worker) {
	logger.SetLevel(conf.LogLevel)

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems?
	// Configure logging
	if conf.LogPath != "" {
		logFile, err := os.OpenFile(conf.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Can't open log output file", "path", conf.LogPath)
		} else {
			logger.SetOutput(logFile)
		}
	}

	if conf.ID == "" {
		conf.ID = scheduler.GenWorkerID("funnel")
	}

	w, err := worker.NewWorker(conf)
	if err != nil {
		log.Error("Can't create worker", err)
		return
	}
	w.Run()
}
