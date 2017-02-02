package logger

import (
	"github.com/Sirupsen/logrus"
	"io"
)

func init() {
	logrus.SetFormatter(&textFormatter{
		DisableTimestamp: true,
	})
	// TODO hard-coded level
	logrus.SetLevel(logrus.DebugLevel)
}

// Logger is repsonsible for logging messages from code.
type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Error(string, ...interface{})
  WithFields(...interface{}) Logger
}

// New returns a new Logger instance.
func New(ns string, args ...interface{}) Logger {
	f := fields(args...)
	f["ns"] = ns
  l := logrus.WithFields(f)
	return &logger{l}
}

type logger struct {
  log *logrus.Entry
}

func fields(args ...interface{}) map[string]interface{} {
	f := make(map[string]interface{}, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		k := args[i].(string)
		v := args[i+1]
		f[k] = v
	}
	if len(args)%2 != 0 {
		f["unknown"] = args[len(args)-1]
	}
	return f
}

// Debug logs a debug message.
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Debug("Some message here", "key1", value1, "key2", value2)
func (l *logger) Debug(msg string, args ...interface{}) {
	f := fields(args...)
	logrus.WithFields(f).Debug(msg)
}

// Info logs an info message
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Info("Some message here", "key1", value1, "key2", value2)
func (l *logger) Info(msg string, args ...interface{}) {
	f := fields(args...)
	logrus.WithFields(f).Info(msg)
}

// Error logs an error message
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Error("Some message here", "key1", value1, "key2", value2)
//
// Error has a two-argument version that can be used as a shortcut.
//     err := startServer()
//     log.Error("Couldn't start server", err)
func (l *logger) Error(msg string, args ...interface{}) {
	var f map[string]interface{}
	if len(args) == 1 {
		f = fields("error", args[0])
	} else {
		f = fields(args...)
	}
	logrus.WithFields(f).Error(msg)
}

// WithFields returns a new Logger instance with the given fields added to all log messages.
func (l *logger) WithFields(args ...interface{}) Logger {
  f := fields(args...)
  n := l.log.WithFields(f)
  return &logger{n}
}

// SetOutput sets the output for all loggers.
func SetOutput(w io.Writer) {
	logrus.SetOutput(w)
}

var rootLogger = New("tes")

// Debug logs to the global logger at the Debug level
func Debug(msg string, args ...interface{}) {
	rootLogger.Debug(msg, args...)
}

// Info logs to the global logger at the Info level
func Info(msg string, args ...interface{}) {
	rootLogger.Info(msg, args...)
}

// Error logs to the global logger at the Error level
func Error(msg string, args ...interface{}) {
	rootLogger.Error(msg, args...)
}
