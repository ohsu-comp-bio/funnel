package events

import (
	"context"
)

// Writer provides write access to a task's events
type Writer interface {
	WriteEvent(context.Context, *Event) error
}

// MultiWriter allows writing an event to a list of Writers.
// Writing stops on the first error.
type MultiWriter []Writer

// Append appends the given Writers to the MultiWriter.
func (mw *MultiWriter) Append(ws ...Writer) {
	*mw = append(*mw, ws...)
}

// WriteEvent writes an event to all the writers. Writing stops on the first error.
func (mw *MultiWriter) WriteEvent(ctx context.Context, ev *Event) error {
	for _, w := range *mw {
		err := w.WriteEvent(ctx, ev)
		if err != nil {
			return err
		}
	}
	return nil
}
