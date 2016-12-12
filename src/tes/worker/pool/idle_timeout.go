package pool

import "time"

// IdleTimer provides a helper for managing the Pool's idle timeout.
// The timer will write a value to the Done channel when the idle duration has elapsed.
//
// See Pool.Start() in pool.go for an example.
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
	return
}
func (*noTimeout) Stop() {
	return
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

func (t *timerTimeout) Done() <-chan time.Time {
	if t.timer != nil {
		return t.timer.C
	}
	return nil
}

func (t *timerTimeout) Start() {
	if !t.started {
		t.timer = time.NewTimer(t.timeout)
		t.started = true
	}
}

func (t *timerTimeout) Stop() {
	if !t.timer.Stop() {
		// If the timer already finished, drain the channel.
		<-t.timer.C
	}
	t.started = false
}

func IdleTimeoutAfter(d time.Duration) IdleTimeout {
	return &timerTimeout{d, nil, false}
}

func IdleTimeoutAfterSeconds(sec int) IdleTimeout {
	return IdleTimeoutAfter(time.Duration(sec) * time.Second)
}
