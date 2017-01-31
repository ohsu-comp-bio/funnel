// Package slot contains concepts that organize concurrent job processing in a single worker.
// A slot can process one job at a time. A worker can have a pool of slots.
// A slot is responsible for requesting jobs from the scheduler and processing them.
package slot

import (
	"context"
	"fmt"
	"sync"
  "tes/logger"
	"tes/scheduler"
	worker "tes/worker"
)

// State represents the state of a slot (e.g. Idle, Running, etc).
type State int32

const (
	// Idle means the slot is waiting for a job from the scheduler.
	Idle State = iota
	// Running means the slots is currently running a job.
	Running
)

// Slot is responsible for requesting a job from the scheduler, running it,
// and repeating.
type Slot struct {
	ID       string
	sched    *scheduler.Client
	engine   worker.Engine
	state    State
	stateMtx sync.Mutex
  log      logger.Logger
}

// NewSlot returns a new Slot instance.
func NewSlot(id string, schedAddr string, engine worker.Engine) (*Slot, error) {

	// Get a client for the scheduler service
	sched, err := scheduler.NewClient(schedAddr)
	if err != nil {
		return nil, err
	}

  log := logger.New("slot", "slotID", id)

	return &Slot{
		ID:     id,
		sched:  sched,
		engine: engine,
    log:    log,
	}, nil
}

// Close closes the Slot and cleans up resources.
func (slot *Slot) Close() {
	slot.sched.Close()
}

// State gets the state of the slot: either running or idle.
// This helps track the state of a pool of slots, so it can shutdown if idle.
func (slot *Slot) State() (state State) {
	slot.stateMtx.Lock()
	state = slot.state
	slot.stateMtx.Unlock()
	return
}

// setState sets the state of the slot (to either running or idle.)
func (slot *Slot) setState(state State) {
	// Slots are currently used across goroutines, so this requires thread-safety via a mutex lock.
	slot.stateMtx.Lock()
	slot.state = state
	slot.stateMtx.Unlock()
}

// Run starts job processing. Ask the scheduler for a job, run it,
// send state updates to the pool/log/scheduler, and repeat.
func (slot *Slot) Run(ctx context.Context) {
	slot.log.Info("Starting")

	for {
		select {
		case <-ctx.Done():
			// The context was canceled (maybe the slot is being shut down) so return.
      slot.log.Info("Done")
			return
		default:
			// This blocks until a job is available, or the context is canceled.
			// It's possible to return nil (if the context is canceled), so we
			// have to check the return value below.
			job := slot.sched.PollForJob(ctx, slot.ID)
			if job != nil {
				// Set the slot state to running
				slot.setState(Running)
				// This blocks until the job is finished.
				slot.engine.RunJob(ctx, job)
			}
			// Set the slot state to idle
			slot.setState(Idle)
		}
	}
}

// GenSlotID generates a new ID for a slot, based on the given pool ID + slot number.
func GenSlotID(id PoolID, i int) string {
	return fmt.Sprintf("%s-%d", id, i)
}
