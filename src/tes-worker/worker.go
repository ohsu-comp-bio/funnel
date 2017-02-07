package main

import (
	"flag"
	"os"
	"tes/config"
	"tes/logger"
	worker "tes/worker"
	"tes/worker/slot"
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
	flag.IntVar(&conf.Slots, "num-slots", conf.Slots, "Worker Slot Count")
	flag.StringVar(&conf.LogPath, "log-path", conf.LogPath, "File path to write logs to")
	flag.Parse()

	config.LoadConfigOrExit(confArg, &conf)
	logger.SetLevel(conf.LogLevel)
	start(conf)
}

func start(conf config.Worker) {

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

	// Create the job engine
	eng, err := worker.NewEngine(conf)
	if err != nil {
		log.Error("Couldn't create engine", err)
		return
	}

	// Configure the slot timeout
	idleTimeout := slot.NoIdleTimeout()
	if conf.Timeout != -1 {
		idleTimeout = slot.IdleTimeoutAfterSeconds(conf.Timeout)
	}

	// Create the slot pool
	slots := make([]*slot.Slot, conf.Slots)
	p := slot.NewPool(slots, idleTimeout)

	// Create the slots
	for i := 0; i < conf.Slots; i++ {
		// TODO handle error
		slots[i], _ = slot.NewSlot(conf, eng)
	}

	// Start the pool
	p.Start()
}
