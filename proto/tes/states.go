package tes

import "fmt"

// State variables for convenience
const (
	Unknown       = State_UNKNOWN
	Queued        = State_QUEUED
	Running       = State_RUNNING
	Paused        = State_PAUSED
	Complete      = State_COMPLETE
	Error         = State_ERROR
	Canceled      = State_CANCELED
	Initializing  = State_INITIALIZING
)

func transitionError(from, to State) error {
	return fmt.Errorf("invalid state transition from %s to %s",
		from.String(), to.String())
}

// ValidateTransition validates a task state transition.
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
			return transitionError(from, to)
		}
		return nil

	case Initializing:

		switch to {
		case Unknown, Queued:
			return transitionError(from, to)
		case Running, Error, Canceled:
			return nil
		}

	case Running:

		switch to {
		case Unknown, Queued:
			return transitionError(from, to)
		case Complete, Error, Canceled:
			return nil
		}

	case Error, Canceled, Complete:
		// May not transition out of terminal state.
		return transitionError(from, to)

	default:
		return fmt.Errorf("unknown state: %s", from)
	}
	// Shouldn't be reaching this point, but just in case.
	return transitionError(from, to)
}
