package events

import (
	"github.com/ohsu-comp-bio/funnel/logger"
)

// ErrLogger writes an error message to the given logger when an event write fails.
type ErrLogger struct {
	Writer
	Log *logger.Logger
}

func (e *ErrLogger) Write(ev *Event) error {
	err := e.Writer.Write(ev)
	if err != nil {
		e.Log.Error("error writing event", err)
	}
	return err
}
