// slot organizes concurrent job processing
package slot

import (
	"context"
	uuid "github.com/nu7hatch/gouuid"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tes/worker"
	"time"
)

// For brevity
type Context context.Context

// Used to pass file system configuration to the worker engine.
type FileConfig struct {
	SwiftCacheDir string
	AllowedDirs   string
	SharedDir     string
	VolumeDir     string
}

type PoolId string

// Pool provides concurrent job processing.
// The pool has a number of "slots", which determines how many jobs are processed concurrently.
//
// If the pool/slots sit idle for longer than Pool.idleTimeout, the pool will exit.
type Pool struct {
	Id PoolId
	// sleepDuration controls how often the pool will check the status of its slots.
	statusCheckDuration time.Duration
	// idleTimeout controls how long before the pool shuts down when no jobs are available.
	idleTimeout IdleTimeout
	slots       []*Slot
}

// Create a Pool instance with basic default parameters.
func NewPool(slots []*Slot, idleTimeout IdleTimeout) *Pool {
	return &Pool{
		Id:                  GenPoolId(),
		statusCheckDuration: time.Second * 2,
		idleTimeout:         idleTimeout,
		slots:               slots,
	}
}

// Start the slots and track their status.
// If the pool is idle for longer than Pool.idleTimeout, exit.
func (pool *Pool) Start() {
	ctx := context.Background()
	defer ctx.Done()
	defer log.Printf("Shutting down pool")

	ticker := time.NewTicker(pool.statusCheckDuration)
	defer ticker.Stop()

	log.Printf("Starting pool of %d slots", len(pool.slots))

	// Create and start slots.
	for _, slot := range pool.slots {
		go slot.Start(ctx)
	}

	// This tracks the status of concurrent job slots.
	// If no jobs are found for awhile, the pool will shut down.
	for {
		select {
		case <-pool.idleTimeout.Done():
			// Break the loop for shutdown
			log.Printf("Pool has reached idle timeout")
			return
		case <-ticker.C:
			// Check if the pool is completely idle.
			isRunning := false
			for _, slot := range pool.slots {
				if slot.Status() == Running {
					isRunning = true
				}
			}

			if isRunning {
				pool.idleTimeout.Stop()
			} else {
				pool.idleTimeout.Start()
			}
			// TODO what if there are active slots, but they aren't sending status updates?
		}
	}
}

// Generate an ID for the a Pool.
// Currently, this generates a UUID string.
func GenPoolId() PoolId {
	u, _ := uuid.NewV4()
	return PoolId(u.String())
}

// TODO I'm not sure what the best place for this is, or how/when is best to create
//      the file mapper/client.
func getFileClient(config FileConfig) tesTaskEngineWorker.FileSystemAccess {

	if config.VolumeDir != "" {
		volumeDir, _ := filepath.Abs(config.VolumeDir)
		if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
			os.Mkdir(volumeDir, 0700)
		}
	}

	// OpenStack Swift object storage
	if config.SwiftCacheDir != "" {
		// Mock Swift storage directory to local filesystem.
		// NOT actual swift.
		storageDir, _ := filepath.Abs(config.SwiftCacheDir)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}

		return tesTaskEngineWorker.NewSwiftAccess()

		// Local filesystem storage
	} else if config.AllowedDirs != "" {
		o := []string{}
		for _, i := range strings.Split(config.AllowedDirs, ",") {
			p, _ := filepath.Abs(i)
			o = append(o, p)
		}
		return tesTaskEngineWorker.NewFileAccess(o)

		// Shared filesystem storage
	} else if config.SharedDir != "" {
		storageDir, _ := filepath.Abs(config.SharedDir)
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			os.Mkdir(storageDir, 0700)
		}
		return tesTaskEngineWorker.NewSharedFS(storageDir)

	} else {
		// TODO what's a good default? Or error?
		return tesTaskEngineWorker.NewSharedFS("storage")
	}
}
