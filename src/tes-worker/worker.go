package main

import (
	"flag"
	"log"
	"os"
	worker "tes/worker"
	"tes/worker/slot"
)

type WorkerConfig struct {
	masterAddr string
	fileConfig worker.FileConfig
	volumeDir  string
	timeout    int
	numWorkers int
	logPath    string
}

func main() {
	masterArg := flag.String("master", "localhost:9090", "Master Server")
	volumeDirArg := flag.String("volumes", "volumes", "Volume Dir")
	timeoutArg := flag.Int("timeout", -1, "Timeout in seconds")
	nworker := flag.Int("nworkers", 4, "Worker Count")
	logFileArg := flag.String("logfile", "", "File path to write logs to")

	flag.Parse()

	config := WorkerConfig{
		masterAddr: *masterArg,
		volumeDir:  *volumeDirArg,
		timeout:    *timeoutArg,
		numWorkers: *nworker,
		logPath:    *logFileArg,
	}
	start(config)
}

func start(config WorkerConfig) {
	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems?
	if config.logPath != "" {
		logFile, err := os.OpenFile(config.logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Can't open log file")
		} else {
			log.SetOutput(logFile)
		}
	}

	idleTimeout := slot.NoIdleTimeout()
	if config.timeout != -1 {
		idleTimeout = slot.IdleTimeoutAfterSeconds(config.timeout)
	}

	// TODO config changed. what's the right thing to do here?
	fileConfig := worker.FileConfig{
		VolumeDir: config.volumeDir,
	}

	slots := make([]*slot.Slot, config.numWorkers)
	p := slot.NewPool(slots, idleTimeout)

	eng, err := worker.NewEngine(config.masterAddr, fileConfig)
	if err != nil {
		log.Printf("Error creating worker engine: %s", err)
		return
	}

	for i := 0; i < config.numWorkers; i++ {
		id := slot.GenSlotId(p.Id, i)
		// TODO handle error
		slots[i], _ = slot.NewSlot(id, config.masterAddr, eng)
	}

	p.Start()
}
