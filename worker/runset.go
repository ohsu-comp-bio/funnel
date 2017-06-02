package worker

import (
	"context"
	"sync"
)

// runSet tracks a set of concurrent goroutines by ID.
// Used by the worker service to track a set of running tasks by task ID.
type runSet struct {
	wg      sync.WaitGroup
	mtx     sync.Mutex
	runners map[string]context.CancelFunc
}

// Run will call the "run" function in a gouroutine and increment the waitgroup count.
// Ensures "run" is only called once per ID.
func (r *runSet) Add(id string, run func(context.Context, string)) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Initialize map if needed
	if r.runners == nil {
		r.runners = make(map[string]context.CancelFunc)
	}

	// If there's already a runner for the given task ID,
	// do nothing.
	if _, ok := r.runners[id]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.runners[id] = cancel

	r.wg.Add(1)
	go func() {
		run(ctx, id)
		r.wg.Done()

		// When the task is finished, remove the task ID from the set
		r.mtx.Lock()
		defer r.mtx.Unlock()
		delete(r.runners, id)
	}()
}

// Cancel all runners and wait for them to exit.
func (r *runSet) Stop() {
	for _, cancel := range r.runners {
		cancel()
	}
	r.runners = nil
	r.wg.Wait()
}

// Wait for all runners to exit.
func (r *runSet) Wait() {
	r.wg.Wait()
}

// Count returns the number of runners currently running.
func (r *runSet) Count() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return len(r.runners)
}
