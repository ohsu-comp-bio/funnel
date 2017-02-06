package main

import (
	"flag"
	"os"
	"tes"
	"tes/logger"
	worker "tes/worker"
	"tes/worker/slot"
)

var log = logger.New("tes-worker")

func main() {
	config := tes.WorkerDefaultConfig()

	var configArg string
	flag.StringVar(&configArg, "config", "", "Config File")
	flag.StringVar(&config.ID, "id", config.ID, "Worker ID")
	flag.StringVar(&config.ServerAddress, "server-address", config.ServerAddress, "Server address")
	flag.StringVar(&config.WorkDir, "work-dir", config.WorkDir, "Working Directory")
	flag.IntVar(&config.Timeout, "timeout", config.Timeout, "Timeout in seconds")
	flag.IntVar(&config.Slots, "num-slots", config.Slots, "Worker Slot Count")
	flag.StringVar(&config.LogPath, "log-path", config.LogPath, "File path to write logs to")
	flag.Parse()

	tes.LoadConfigOrExit(configArg, &config)
	start(config)
}

func start(config tes.Worker) {

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems?
	// Configure logging
	if config.LogPath != "" {
		logFile, err := os.OpenFile(config.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Can't open log output file", "path", config.LogPath)
		} else {
			logger.SetOutput(logFile)
		}
	}

	// Create the job engine
	eng, err := worker.NewEngine(config)
	if err != nil {
		log.Error("Couldn't create engine", err)
		return
	}

	// Configure the slot timeout
	idleTimeout := slot.NoIdleTimeout()
	if config.Timeout != -1 {
		idleTimeout = slot.IdleTimeoutAfterSeconds(config.Timeout)
	}

	// Create the slot pool
	slots := make([]*slot.Slot, config.Slots)
	p := slot.NewPool(slots, idleTimeout)

	// Create the slots
	for i := 0; i < config.Slots; i++ {
		// TODO handle error
		slots[i], _ = slot.NewSlot(config, eng)
	}

	// Start the pool
	p.Start()
}
