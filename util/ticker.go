package util

import (
	"context"
	"time"
)

// Ticker is a wrapper around time.Ticker which
// 1) fires immediately
// 2) can be canceled by the given context.
func Ticker(ctx context.Context, d time.Duration) <-chan time.Time {
	// This code is likely trickier than you expect.

	// Output channel.
	out := make(chan time.Time)
	done := make(chan struct{})

	ticker := time.NewTicker(d)
	go func() {
		<-ctx.Done()
		ticker.Stop()
		close(done)
	}()

	go func() {
		defer close(out)
		// Fire tick immediately
		out <- time.Now()

		for {
			select {
			case t := <-ticker.C:
				out <- t
			case <-done:
				return
			}
		}
	}()
	return out
}
