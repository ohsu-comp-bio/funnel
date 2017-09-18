package util

import (
	"context"
	"os"
	"os/signal"
)

// SignalContext will cancel the context when any of the given
// signals is received.
func SignalContext(ctx context.Context, sigs ...os.Signal) context.Context {
	sch := make(chan os.Signal, 1)
	sub, cancel := context.WithCancel(ctx)

	signal.Notify(sch, sigs...)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sch:
			cancel()
			return
		}
	}()
	return sub
}
