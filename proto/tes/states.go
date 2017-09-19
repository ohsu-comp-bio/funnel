package tes

import (
	"fmt"
)

// State variables for convenience
const (
	Unknown      = State_UNKNOWN
	Queued       = State_QUEUED
	Running      = State_RUNNING
	Paused       = State_PAUSED
	Complete     = State_COMPLETE
	Error        = State_ERROR
	SystemError  = State_SYSTEM_ERROR
	Canceled     = State_CANCELED
	Initializing = State_INITIALIZING
)

// Transitioner defines the interface for handling task state transitions.
type Transitioner interface {
	// The implementation should remove the task from the queue and set the state given by "to".
	Dequeue(to State) error
	// The implementation should put an existing task back in the queue (used for restarts).
	Requeue() error
	// The implementation should set the state of the task.
	SetState(to State) error
}

// Transition validates a task state transition and, if valid, calls the corresponding
// function on the given Transitioner. Valid transitions are:
//
// - Queued -> Initializing, Running, Canceled
func Transition(from, to State, t Transitioner) error {

	if from == to {
		return nil
	}

	switch from {
	case Queued:

		switch to {
		case Initializing, Running, Canceled,
			SystemError, Error:
			return t.Dequeue(to)
		}

	case Initializing:

		switch to {
		case Running, Error, SystemError, Canceled:
			return t.SetState(to)
		}

	case Running:

		switch to {
		case Complete, Error, SystemError, Canceled:
			return t.SetState(to)
		}

	case Error, SystemError, Canceled:

		if to == Queued {
			return t.Requeue()
		}

	case Paused:
		return fmt.Errorf("Paused state is not implemented.")

	default:
		return fmt.Errorf("Unknown state: %s", from)
	}
	return fmt.Errorf("Unhandled state transition from %s to %s",
		from, to)
}
