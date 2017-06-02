package tes

func RunnableState(s State) bool {
	return s == State_INITIALIZING || s == State_RUNNING
}

func TerminalState(s State) bool {
	return s == State_COMPLETE || s == State_ERROR || s == State_SYSTEM_ERROR ||
		s == State_CANCELED
}
