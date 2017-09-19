package events

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// EventLogger writes events to a funnel logger.
type EventLogger struct {
	Log logger.Logger
}

// NewEventLogger creates an event logger and a funnel logger with the given name.
func NewEventLogger(name string) *EventLogger {
	return &EventLogger{logger.Sub(name)}
}

// Write writes an event to the logger.
func (el *EventLogger) Write(ev *Event) error {
	ts := ev.Type.String()
	log := el.Log.WithFields("taskID", ev.Id, "attempt", ev.Attempt,
		"timestamp", ptypes.TimestampString(ev.Timestamp))

	switch ev.Type {
	case Type_STATE:
		log.Info(ts, "state", ev.State.String())
	case Type_START_TIME:
		log.Info(ts, "start_time", ptypes.TimestampString(ev.StartTime))
	case Type_END_TIME:
		log.Info(ts, "end_time", ptypes.TimestampString(ev.EndTime))
	case Type_OUTPUTS:
		log.Info(ts, "outputs", ev.Outputs)
	case Type_METADATA:
		log.Info(ts, "metadata", ev.Metadata)
	case Type_EXECUTOR_START_TIME:
		log.Info(ts, "start_time", ptypes.TimestampString(ev.ExecutorStartTime))
	case Type_EXECUTOR_END_TIME:
		log.Info(ts, "end_time", ptypes.TimestampString(ev.ExecutorEndTime))
	case Type_EXIT_CODE:
		log.Info(ts, "exit_code", ev.ExitCode)
	case Type_HOST_IP:
		log.Info(ts, "host_ip", ev.HostIp)
	case Type_PORTS:
		log.Info(ts, "ports", ev.Ports)
	case Type_STDOUT:
		log.Info(ts, "stdout", ev.Stdout)
	case Type_STDERR:
		log.Info(ts, "stderr", ev.Stderr)
	case Type_SYSLOG:
		var args []interface{}
		for k, v := range ev.SystemLog.Fields {
			args = append(args, k, v)
		}
		switch ev.SystemLog.Level {
		case "error":
			log.Error(ev.SystemLog.Msg, args...)
		case "info":
			log.Info(ev.SystemLog.Msg, args...)
		case "debug":
			log.Debug(ev.SystemLog.Msg, args...)
		}
	default:
		log.Info(ts, "event", ev)
	}
	return nil
}

type multiwriter []Writer

// MultiWriter writes events to all the given writers.
func MultiWriter(ws ...Writer) Writer {
	return multiwriter(ws)
}

// Write writes an event to all the writers.
func (mw multiwriter) Write(ev *Event) error {
	for _, w := range mw {
		err := w.Write(ev)
		if err != nil {
			return err
		}
	}
	return nil
}
