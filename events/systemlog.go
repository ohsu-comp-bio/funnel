package events

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/util"
)

// SystemLogGenerator is a type that emulates the logger interface
// and outputs Events.
type SystemLogGenerator struct {
	taskID  string
	attempt uint32
	index   uint32
}

// Info creates an Event for an 'info' level log message
func (sle *SystemLogGenerator) Info(msg string, args ...interface{}) *Event {
	return NewSystemLog(sle.taskID, sle.attempt, sle.index, "info", msg, fields(args...))
}

// Debug creates an Event for a 'debug' level log message
func (sle *SystemLogGenerator) Debug(msg string, args ...interface{}) *Event {
	return NewSystemLog(sle.taskID, sle.attempt, sle.index, "debug", msg, fields(args...))
}

// Error creates an Event for an 'error' level log message
func (sle *SystemLogGenerator) Error(msg string, args ...interface{}) *Event {
	return NewSystemLog(sle.taskID, sle.attempt, sle.index, "error", msg, fields(args...))
}

// SystemLogWriter is a type that emulates the logger interface
// and outputs Events.
type SystemLogWriter struct {
	lvl string
	gen *SystemLogGenerator
	out Writer
}

// Info writes an Event for an 'info' level log message
func (sle *SystemLogWriter) Info(msg string, args ...interface{}) error {
	if sle.lvl != "error" {
		return sle.out.WriteEvent(context.Background(), sle.gen.Info(msg, args...))
	}
	return nil
}

// Debug writes an  for a 'debug' level log message
func (sle *SystemLogWriter) Debug(msg string, args ...interface{}) error {
	if sle.lvl == "debug" {
		return sle.out.WriteEvent(context.Background(), sle.gen.Debug(msg, args...))
	}
	return nil
}

// Error writes an Event for an 'error' level log message
func (sle *SystemLogWriter) Error(msg string, args ...interface{}) error {
	return sle.out.WriteEvent(context.Background(), sle.gen.Error(msg, args...))
}

// converts an argument list to a map, e.g.
// ("key", value, "key2", value2) => {"key": value, "key2", value2}
func fields(args ...interface{}) map[string]string {
	ss := make(map[string]string)
	si := make(map[string]interface{})
	si = util.ArgListToMap(args...)
	for k, v := range si {
		ss[k] = fmt.Sprintf("%+v", v)
	}
	return ss
}
