package slot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"tes/scheduler"
	worker "tes/worker"
)

// Used by slots to communicate their state with the pool.
type SlotState int32

const (
	Idle SlotState = iota
	Running
)

// A slot requests a job from the scheduler, runs it, and repeats.
// There are many slots in a pool, which provides job concurrency.
type Slot struct {
	Id       string
	sched    *scheduler.Client
	engine   worker.Engine
	state    SlotState
	stateMtx sync.Mutex
}

func NewSlot(id string, schedAddr string, engine worker.Engine) (*Slot, error) {

	// Get a client for the scheduler service
	sched, err := scheduler.NewClient(schedAddr)
	if err != nil {
		return nil, err
	}

	return &Slot{
		Id:     id,
		sched:  sched,
		engine: engine,
	}, nil
}

func (slot *Slot) Close() {
	slot.sched.Close()
}

// State gets the state of the slot: either running or idle.
// This helps track the state of a pool of slots, so it can shutdown if idle.
func (slot *Slot) State() (state SlotState) {
	slot.stateMtx.Lock()
	state = slot.state
	slot.stateMtx.Unlock()
	return
}

// setState sets the state of the slot (to either running or idle.)
func (slot *Slot) setState(state SlotState) {
	// Slots are currently used across goroutines, so this requires thread-safety via a mutex lock.
	slot.stateMtx.Lock()
	slot.state = state
	slot.stateMtx.Unlock()
}

// Run starts job processing. Ask the scheduler for a job, run it,
// send state updates to the pool/log/scheduler, and repeat.
func (slot *Slot) Run(ctx context.Context) {
	log.Printf("Starting slot: %s", slot.Id)

	for {
		select {
		case <-ctx.Done():
			// The context was canceled (maybe the slot is being shut down) so return.
			log.Printf("Slot is done: %s", slot.Id)
			return
		default:
			// This blocks until a job is available, or the context is canceled.
			// It's possible to return nil (if the context is canceled), so we
			// have to check the retun value below.
			job := slot.sched.PollForJob(ctx, slot.Id)
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

// Generate a slot ID based on a pool ID + slot number.
func GenSlotId(poolId PoolId, i int) string {
	return fmt.Sprintf("%s-%d", poolId, i)
}
