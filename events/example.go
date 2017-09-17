package events

// example event type switch code for easy copy/paste reuse
func exampleSwitch(ev *Event) {
	switch ev.Type {
	case Type_STATE:
	case Type_START_TIME:
	case Type_END_TIME:
	case Type_OUTPUTS:
	case Type_METADATA:
	case Type_EXECUTOR_START_TIME:
	case Type_EXECUTOR_END_TIME:
	case Type_EXIT_CODE:
	case Type_HOST_IP:
	case Type_PORTS:
	case Type_STDOUT:
	case Type_STDERR:
	case Type_SYSLOG:

		for range ev.SystemLog.Fields {
		}

		switch ev.SystemLog.Level {
		case "error":
		case "info":
		case "debug":
		}
	default:
	}
}
