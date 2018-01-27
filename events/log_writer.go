package events

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// Logger writes events to a funnel logger.
type Logger struct {
	Log *logger.Logger
}

// WriteEvent writes an event to the logger.
func (el *Logger) WriteEvent(ctx context.Context, ev *Event) error {
	ts := ev.Type.String()
	log := el.Log.WithFields(
		"taskID", ev.Id,
		"timestamp", ev.Timestamp,
	)

	switch ev.Type {
	case Type_STATE:
		log.Info(ts, "state", ev.GetState().String())
	case Type_START_TIME:
		log.Info(ts, "start_time", ev.GetStartTime())
	case Type_END_TIME:
		log.Info(ts, "end_time", ev.GetEndTime())
	case Type_OUTPUTS:
		log.Info(ts, "outputs", ev.GetOutputs().Value)
	case Type_METADATA:
		log.Info(ts, "metadata", ev.GetMetadata().Value)
	case Type_EXIT_CODE:
		log.Info(ts, "exit_code", ev.GetExitCode())
	case Type_STDOUT:
		log.Info(ts, "stdout", ev.GetStdout())
	case Type_STDERR:
		log.Info(ts, "stderr", ev.GetStderr())
	case Type_SYSTEM_LOG:
		var args []interface{}
		for k, v := range ev.GetSystemLog().Fields {
			args = append(args, k, v)
		}
		switch ev.GetSystemLog().Level {
		case "warning":
			log.Warn(ev.GetSystemLog().Msg, args...)
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
