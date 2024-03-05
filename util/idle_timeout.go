package util

import "time"

// IdleTimeout provides a helper for managing the Pool's idle timeout.
// Start() and Stop() are used to control the timer, and Done() is used to
// detect when the timeout has been reached.
//
// in := make(chan int)
// requestInput(in)
// t := IdleTimeoutAfter(time.Second * 10)
//
//	for {
//	  select {
//	  case <-t.Done():
//	    // ... code to respond to timeout
//	  case <-in:
//	    // Reset the timeout.
//	    t.Start()
//	  }
//	}
type IdleTimeout interface {
	Done() <-chan time.Time
	Start()
	Stop()
}

// noTimeout never times out.
// It has no fields and doesn't need to do anything except provide the IdleTimeout interface.
type noTimeout struct{}

func (*noTimeout) Done() <-chan time.Time {
	return nil
}
func (*noTimeout) Start() {
	// return
}
func (*noTimeout) Stop() {
	// return
}

// NoIdleTimeout creates an IdleTimeout that never times out.
func NoIdleTimeout() IdleTimeout {
	return new(noTimeout)
}

type timerTimeout struct {
	timeout time.Duration
	timer   *time.Timer
	started bool
}

// Done returns a channel which can be used to wait for the timeout.
//
//	t := IdleTimeoutAfter(time.Second * 10)
//	for {
//	  select {
//	  case <-t.Done():
//	    // ... code to respond to timeout
//	  }
//	}
func (t *timerTimeout) Done() <-chan time.Time {
	if t.timer != nil {
		return t.timer.C
	}
	return nil
}

// Start resets and starts the timer.
func (t *timerTimeout) Start() {
	if !t.started {
		t.timer = time.NewTimer(t.timeout)
		t.started = true
	}
}

// Stop stops the idle timer and cleans up resources.
func (t *timerTimeout) Stop() {
	if !t.started {
		return
	}
	t.timer.Stop()
	t.started = false
}

// NewIdleTimeout returns an IdleTimeout instance for the given duration.
func NewIdleTimeout(d time.Duration) IdleTimeout {
	if d < 0 {
		return NoIdleTimeout()
	}
	return &timerTimeout{d, nil, false}
}
