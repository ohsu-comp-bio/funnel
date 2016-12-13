// slot organizes concurrent job processing
package slot

import (
	"context"
	uuid "github.com/nu7hatch/gouuid"
	"log"
	"time"
)

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
		go slot.Run(ctx)
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
}

// Generate an ID for the a Pool.
// Currently, this generates a UUID string.
func GenPoolId() PoolId {
	u, _ := uuid.NewV4()
	return PoolId(u.String())
}
