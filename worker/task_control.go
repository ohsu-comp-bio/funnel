package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"sync"
)

// State variables for convenience
const (
	Unknown      = tes.State_UNKNOWN
	Queued       = tes.State_QUEUED
	Running      = tes.State_RUNNING
	Paused       = tes.State_PAUSED
	Complete     = tes.State_COMPLETE
	Error        = tes.State_ERROR
	SystemError  = tes.State_SYSTEM_ERROR
	Canceled     = tes.State_CANCELED
	Initializing = tes.State_INITIALIZING
)

// TaskState represents the state of a running task
type TaskState interface {
	Err() error
	State() tes.State
	Complete() bool
}

// TaskControl represents control over a running task
type TaskControl interface {
	TaskState
	Cancel()
	SetRunning()
	SetResult(error)
	Context() context.Context
}

// NewTaskControl returns a new TaskControl instance
func NewTaskControl() TaskControl {
	ctx, cancel := context.WithCancel(context.Background())
	return &taskControl{ctx: ctx, cancelFunc: cancel}
}

type taskControl struct {
	running    bool
	complete   bool
	err        error
	mtx        sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (r *taskControl) Context() context.Context {
	return r.ctx
}

func (r *taskControl) SetResult(err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	// Don't set the result twice
	if !r.complete {
		r.complete = true
		r.err = err
	}
}

func (r *taskControl) SetRunning() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.complete {
		r.running = true
	}
}

func (r *taskControl) Err() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.err
}

func (r *taskControl) Cancel() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.cancelFunc()
	r.err = r.ctx.Err()
	r.complete = true
}

func (r *taskControl) Complete() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.complete
}

func (r *taskControl) State() tes.State {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	switch {
	case r.err == context.Canceled:
		return Canceled
	case r.err != nil:
		return Error
	case r.complete:
		return Complete
	case r.running:
		return Running
	default:
		return Initializing
	}
}
