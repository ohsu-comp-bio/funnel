package main

import (
	"flag"
	"os"
	"tes/config"
	"tes/logger"
	"tes/scheduler"
	"tes/worker"
)

var log = logger.New("tes-worker")

func main() {
	conf := config.WorkerDefaultConfig()

	var confArg string
	flag.StringVar(&confArg, "config", "", "Config File")
	flag.StringVar(&conf.ID, "id", conf.ID, "Worker ID")
	flag.StringVar(&conf.ServerAddress, "server-address", conf.ServerAddress, "Server address")
	flag.StringVar(&conf.WorkDir, "work-dir", conf.WorkDir, "Working Directory")
	flag.DurationVar(&conf.Timeout, "timeout", conf.Timeout, "Timeout in seconds")
	flag.StringVar(&conf.LogPath, "log-path", conf.LogPath, "File path to write logs to")
	flag.Parse()

	config.LoadConfigOrExit(confArg, &conf)
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
		conf.ID = scheduler.GenWorkerID()
	}

	log.Debug("WORKER CONFIG", conf)

	w, err := worker.NewWorker(conf)
	if err != nil {
		log.Error("Can't create worker", err)
		return
	}
	w.Run()
}
