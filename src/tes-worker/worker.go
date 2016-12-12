package main

import (
	"flag"
	"log"
	"os"
	"tes/worker/pool"
)

type WorkerConfig struct {
	masterAddr string
	fileConfig pool.FileConfig
	timeout    int
	numWorkers int
	logPath    string
}

func main() {
	masterArg := flag.String("master", "localhost:9090", "Master Server")
	volumeDirArg := flag.String("volumes", "volumes", "Volume Dir")
	storageDirArg := flag.String("storage", "storage", "Storage Dir")
	allowedDirsArg := flag.String("files", "", "Allowed File Paths")
	swiftDirArg := flag.String("swift", "", "Cache Swift items in directory")
	timeoutArg := flag.Int("timeout", -1, "Timeout in seconds")
	nworker := flag.Int("nworkers", 4, "Worker Count")
	logFileArg := flag.String("logfile", "", "File path to write logs to")

	flag.Parse()

	config := WorkerConfig{
		masterAddr: *masterArg,
		fileConfig: pool.FileConfig{
			SwiftCacheDir: *swiftDirArg,
			AllowedDirs:   *allowedDirsArg,
			SharedDir:     *storageDirArg,
			VolumeDir:     *volumeDirArg,
		},
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

	idleTimeout := pool.NoIdleTimeout()
	if config.timeout != -1 {
		idleTimeout = pool.IdleTimeoutAfterSeconds(config.timeout)
	}

	slots := make([]*pool.Slot, config.numWorkers)
	p := pool.NewPool(slots, idleTimeout)

	for i := 0; i < config.numWorkers; i++ {
		id := pool.GenSlotId(p.Id, i)
		slots[i] = pool.NewDefaultSlot(id, config.masterAddr, config.fileConfig)
	}

	p.Start()
}
