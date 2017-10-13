package events

import (
	"github.com/ohsu-comp-bio/funnel/logger"
)

// Logger writes events to a funnel logger.
type Logger struct {
	Log *logger.Logger
}

// Write writes an event to the logger.
func (el *Logger) Write(ev *Event) error {
	ts := ev.Type.String()
	log := el.Log.WithFields(
		"taskID", ev.Id,
		"attempt", ev.Attempt,
		"index", ev.Index,
		"timestamp", TimestampString(ev.Timestamp),
	)

	switch ev.Type {
	case Type_TASK_STATE:
		log.Info(ts, "state", ev.GetState().String())
	case Type_TASK_START_TIME:
		log.Info(ts, "start_time", TimestampString(ev.GetStartTime()))
	case Type_TASK_END_TIME:
		log.Info(ts, "end_time", TimestampString(ev.GetEndTime()))
	case Type_TASK_OUTPUTS:
		log.Info(ts, "outputs", ev.GetOutputs().Value)
	case Type_TASK_METADATA:
		log.Info(ts, "metadata", ev.GetMetadata().Value)
	case Type_EXECUTOR_START_TIME:
		log.Info(ts, "start_time", TimestampString(ev.GetStartTime()))
	case Type_EXECUTOR_END_TIME:
		log.Info(ts, "end_time", TimestampString(ev.GetEndTime()))
	case Type_EXECUTOR_EXIT_CODE:
		log.Info(ts, "exit_code", ev.GetExitCode())
	case Type_EXECUTOR_HOST_IP:
		log.Info(ts, "host_ip", ev.GetHostIp())
	case Type_EXECUTOR_PORTS:
		log.Info(ts, "ports", ev.GetPorts().Value)
	case Type_EXECUTOR_STDOUT:
		log.Info(ts, "stdout", ev.GetStdout())
	case Type_EXECUTOR_STDERR:
		log.Info(ts, "stderr", ev.GetStderr())
	case Type_SYSTEM_LOG:
		var args []interface{}
		for k, v := range ev.GetSystemLog().Fields {
			args = append(args, k, v)
		}
		switch ev.GetSystemLog().Level {
		case "error":
			log.Error(ev.GetSystemLog().Msg, args...)
		case "info":
			log.Info(ev.GetSystemLog().Msg, args...)
		case "debug":
			log.Debug(ev.GetSystemLog().Msg, args...)
		}
	default:
		log.Info(ts, "event", ev)
	}
	return nil
}

func (el *Logger) Close() error {
	return nil
}
