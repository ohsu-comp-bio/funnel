package slot

import (
	"context"
	uuid "github.com/nu7hatch/gouuid"
	"sync"
	"time"
  "tes/logger"
)

// PoolID is the ID (string) of a pool.
// TODO should probably just drop this type and use string.
type PoolID string

// Pool is a group of Slots, which provides concurrent job processings in a single worker.
// Each slot can process a single job at a time. Pool is fairly lightweight; it mainly
// serves to monitor the slots. Pool can be configured with a timeout; if the slots
// are idle for longer than the timeout, the pool (and therefore worker) will shut down.
type Pool struct {
	ID PoolID
	// sleepDuration controls how often the pool will check the status of its slots.
	statusCheckDuration time.Duration
	// idleTimeout controls how long before the pool shuts down when no jobs are available.
	idleTimeout IdleTimeout
	slots       []*Slot
  log         logger.Logger
}

// NewPool creates a new Pool instance with default configuration.
func NewPool(slots []*Slot, idleTimeout IdleTimeout) *Pool {
  id := GenPoolID()
	return &Pool{
		ID:                  id,
		statusCheckDuration: time.Second * 2,
		idleTimeout:         idleTimeout,
		slots:               slots,
    log:                 logger.New("pool", "poolID", id),
	}
}

// Start starts the slots and monitors their status.
func (pool *Pool) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	// WaitGroup helps wait for the slots to finish and clean up.
	var wg sync.WaitGroup

	defer pool.log.Info("Shutting down")

	// Ticker helps poll slot state
	ticker := time.NewTicker(pool.statusCheckDuration)
	defer ticker.Stop()

  pool.log.Info("Starting", "numSlots", len(pool.slots))

	// Create and start slots.
	// Do some bookkeeping with the WaitGroup and call slot.Run()
	wg.Add(len(pool.slots))
	for _, slot := range pool.slots {
		go func(slot *Slot) {
			defer wg.Done()
			slot.Run(ctx)
		}(slot)
	}

	// This tracks the status of concurrent job slots.
	// If no jobs are found for awhile, the pool will shut down.
loop:
	for {
		select {
		case <-pool.idleTimeout.Done():
			// Break the loop for shutdown
      pool.log.Info("Reached idle timeout")
			break loop
		case <-ticker.C:
			// Check if the pool is completely idle.
			isRunning := false
			for _, slot := range pool.slots {
				if slot.State() == Running {
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

	// Cancel the context, which signals the slots to finish and clean up.
	cancel()
	// Wait for the slots to finish.
	wg.Wait()
}

// GenPoolID generates a new ID (string) for a Pool.
// Currently, this generates a UUID string.
func GenPoolID() PoolID {
	u, _ := uuid.NewV4()
	return PoolID(u.String())
}
