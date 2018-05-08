package tes

import "fmt"

// State variables for convenience
const (
	Unknown       = State_UNKNOWN
	Queued        = State_QUEUED
	Running       = State_RUNNING
	Paused        = State_PAUSED
	Complete      = State_COMPLETE
	ExecutorError = State_EXECUTOR_ERROR
	SystemError   = State_SYSTEM_ERROR
	Canceled      = State_CANCELED
	Initializing  = State_INITIALIZING
)

// TransitionError describes an invalid state transition.
type TransitionError struct {
	From, To State
}

func (te *TransitionError) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s",
		te.From.String(), te.To.String())
}

// ValidateTransition validates a task state transition.
// Returns a TransitionError if the transition is not valid.
func ValidateTransition(from, to State) error {

	if from == to {
		return nil
	}

	if from == Paused || to == Paused {
		return fmt.Errorf("paused state is not implemented")
	}

	switch from {
	case Unknown:
		// May transition from Unknown to anything
		return nil

	case Queued:
		// May transition from Queued to anything except Unknown
		if to == Unknown {
			return &TransitionError{from, to}
		}
		return nil

	case Initializing:

		switch to {
		case Unknown, Queued:
			return &TransitionError{from, to}
		case Running, ExecutorError, SystemError, Canceled:
			return nil
		}

	case Running:

		switch to {
		case Unknown, Queued:
			return &TransitionError{from, to}
		case Complete, ExecutorError, SystemError, Canceled:
			return nil
		}

	case ExecutorError, SystemError, Canceled, Complete:
		// May not transition out of terminal state.
		return &TransitionError{from, to}

	default:
		return &TransitionError{from, to}
	}
	// Shouldn't be reaching this point, but just in case.
	return &TransitionError{from, to}
}
