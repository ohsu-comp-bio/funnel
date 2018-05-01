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

	ticker := time.NewTicker(d)
	go func() {
		<-ctx.Done()
		ticker.Stop()
		close(out)
	}()

	go func() {
		// Fire tick immediately
		out <- time.Now()

		for t := range ticker.C {
			// select + default here prevents sending on a closed channel.
			select {
			case out <- t:
			default:
			}
		}
	}()
	return out
}
