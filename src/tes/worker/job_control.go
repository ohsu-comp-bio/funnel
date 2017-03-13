package worker

import (
	"context"
	"sync"
	pbe "tes/ga4gh"
)

// JobState represents the state of a running job
type JobState interface {
	Err() error
	State() pbe.State
	Complete() bool
}

// JobControl represents control over a running job
type JobControl interface {
	JobState
	Cancel()
	SetRunning()
	SetResult(error)
	Context() context.Context
}

// NewJobControl returns a new JobControl instance
func NewJobControl() JobControl {
	ctx, cancel := context.WithCancel(context.Background())
	return &jobControl{ctx: ctx, cancelFunc: cancel}
}

type jobControl struct {
	running    bool
	complete   bool
	err        error
	mtx        sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (r *jobControl) Context() context.Context {
	return r.ctx
}

func (r *jobControl) SetResult(err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	// Don't set the result twice
	if !r.complete {
		r.complete = true
		r.err = err
	}
}

func (r *jobControl) SetRunning() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if !r.complete {
		r.running = true
	}
}

func (r *jobControl) Err() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.err
}

func (r *jobControl) Cancel() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.cancelFunc()
	r.err = r.ctx.Err()
	r.complete = true
}

func (r *jobControl) Complete() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.complete
}

func (r *jobControl) State() pbe.State {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	switch {
	case r.err == context.Canceled:
		return pbe.State_Canceled
	case r.err != nil:
		return pbe.State_Error
	case r.complete:
		return pbe.State_Complete
	case r.running:
		return pbe.State_Running
	default:
		return pbe.State_Initializing
	}
}
