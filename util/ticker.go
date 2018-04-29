package util

import (
	"context"
	"time"
)

// Ticker is a wrapper around time.Ticker which
// 1) fires immediately
// 2) can be canceled by the given context.
func Ticker(ctx context.Context, d time.Duration) <-chan time.Time {
	out := make(chan time.Time)
	go func() {
		out <- time.Now()
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				out <- t
			}
		}
	}()
	return out
}
