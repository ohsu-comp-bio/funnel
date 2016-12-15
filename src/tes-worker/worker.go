package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tes/storage"
	worker "tes/worker"
	"tes/worker/slot"
)

func main() {
	masterArg := flag.String("master", "localhost:9090", "Master Server")
	workDirArg := flag.String("workdir", "volumes", "Working Directory")
	timeoutArg := flag.Int("timeout", -1, "Timeout in seconds")
	nworker := flag.Int("nworkers", 4, "Worker Count")
	logFileArg := flag.String("logfile", "", "File path to write logs to")

	flag.Parse()

	config := worker.Config{
		MasterAddr: *masterArg,
		WorkDir:    *workDirArg,
		Timeout:    *timeoutArg,
		NumWorkers: *nworker,
		LogPath:    *logFileArg,
	}
	start(config)
}

func start(config worker.Config) {

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems?
	// Configure logging
	if config.LogPath != "" {
		logFile, err := os.OpenFile(config.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Can't open log file")
		} else {
			log.SetOutput(logFile)
		}
	}

	// Configure the storage systems
	allowedDirs := []string{}
	allowedDirs = append(allowedDirs, "/tmp")
	store, _ := new(storage.Storage).WithLocal(allowedDirs)

	// Create the job engine
	eng, err := worker.NewEngine(config.MasterAddr, config.WorkDir, store)
	if err != nil {
		log.Printf("Error creating worker engine: %s", err)
		return
	}

	// Configure the slot timeout
	idleTimeout := slot.NoIdleTimeout()
	if config.Timeout != -1 {
		idleTimeout = slot.IdleTimeoutAfterSeconds(config.Timeout)
	}

	// Create the slot pool
	slots := make([]*slot.Slot, config.NumWorkers)
	p := slot.NewPool(slots, idleTimeout)

	// Create the slots
	for i := 0; i < config.NumWorkers; i++ {
		id := slot.GenSlotId(p.Id, i)
		// TODO handle error
		slots[i], _ = slot.NewSlot(id, config.MasterAddr, eng)
	}

	// Start the pool
	p.Start()
}

func parseAllowedDirs(in string) []string {
	o := []string{}
	for _, i := range strings.Split(in, ",") {
		p, _ := filepath.Abs(i)
		o = append(o, p)
	}
	return o
}
