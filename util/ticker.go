package util

import (
	"context"
	"time"
)

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
