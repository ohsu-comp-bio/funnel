package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tes"
	"tes/storage"
	worker "tes/worker"
	"tes/worker/slot"
)

func main() {
	config := worker.NewConfig()

	var allowedDirsArg csvArg
	var configArg string
	flag.StringVar(&configArg, "config", "", "Config File")
  flag.StringVar(&config.ID, "id", config.ID, "Worker ID")
	flag.StringVar(&config.MasterAddr, "masteraddr", config.MasterAddr, "Master Server")
	flag.StringVar(&config.WorkDir, "workdir", config.WorkDir, "Working Directory")
	flag.IntVar(&config.Timeout, "timeout", config.Timeout, "Timeout in seconds")
	flag.Var(&allowedDirsArg, "alloweddirs", "Allowed directories for local FS backend")
	flag.IntVar(&config.NumWorkers, "numworkers", config.NumWorkers, "Worker Count")
	flag.StringVar(&config.LogPath, "logpath", config.LogPath, "File path to write logs to")

	flag.Parse()

	tes.LoadConfigOrExit(configArg, &config)

	for _, i := range allowedDirsArg {
		p, _ := filepath.Abs(i)
		config.AllowedDirs = append(config.AllowedDirs, p)
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
	store, _ := new(storage.Storage).WithLocal(config.AllowedDirs)

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
		// TODO handle error
		slots[i], _ = slot.NewSlot(config.ID, config.MasterAddr, eng)
	}

	// Start the pool
	p.Start()
}

// csvArg is a helper for allowing comma-separated CLI flags.
// See example 3 here: https://golang.org/src/flag/example_test.go
type csvArg []string

func (a *csvArg) String() string {
	return fmt.Sprint(*a)
}

func (a *csvArg) Set(str string) error {
	for _, i := range strings.Split(str, ",") {
		*a = append(*a, i)
	}
	return nil
}
