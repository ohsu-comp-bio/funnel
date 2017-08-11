package logger

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/logrusorgru/aurora"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Log levels
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	ErrorLevel = "error"
)

// Formatter defines a log output formatter.
type Formatter logrus.Formatter

// Logger is responsible for logging messages from code.
type Logger interface {
	Sub(string, ...interface{}) Logger
	SetFormatter(Formatter)
	SetLevel(string)
	SetOutput(io.Writer)
	Discard()
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Error(string, ...interface{})
	WithFields(...interface{}) Logger
	Configure(Config)
}

// New returns a new Logger instance.
func New(ns string, args ...interface{}) Logger {
	f := fields(args...)
	f["ns"] = ns
	log := logrus.New()
	base := log.WithFields(f)
	l := &logger{log, base}
	l.Configure(DefaultConfig())
	return l
}

type logger struct {
	logrus *logrus.Logger
	base   *logrus.Entry
}

// Sub returns a new sub-logger instance.
func (l *logger) Sub(ns string, args ...interface{}) Logger {
	f := fields(args...)
	f["ns"] = ns
	sl := l.logrus.WithFields(f)
	return &logger{l.logrus, sl}
}

// SetLevel sets the level of the logger.
func (l *logger) SetLevel(lvl string) {
	switch strings.ToLower(lvl) {
	case "debug":
		l.logrus.Level = logrus.DebugLevel
	case "info":
		l.logrus.Level = logrus.InfoLevel
	case "error":
		l.logrus.Level = logrus.ErrorLevel
	default:
		l.logrus.Level = logrus.InfoLevel
	}
}

// SetFormatter sets the formatter of the logger.
func (l *logger) SetFormatter(f Formatter) {
	l.logrus.Formatter = f
}

// SetOutput sets the output of the logger.
func (l *logger) SetOutput(o io.Writer) {
	l.logrus.Out = o
}

// Discard configures the logger to discard all logs.
func (l *logger) Discard() {
	l.SetOutput(ioutil.Discard)
}

// Debug logs a debug message.
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Debug("Some message here", "key1", value1, "key2", value2)
func (l *logger) Debug(msg string, args ...interface{}) {
	defer recoverLogErr()
	f := fields(args...)
	l.base.WithFields(f).Debug(msg)
}

// Info logs an info message
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Info("Some message here", "key1", value1, "key2", value2)
func (l *logger) Info(msg string, args ...interface{}) {
	defer recoverLogErr()
	f := fields(args...)
	l.base.WithFields(f).Info(msg)
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
	defer recoverLogErr()
	f := fields(args...)
	l.base.WithFields(f).Error(msg)
}

// WithFields returns a new Logger instance with the given fields added to all log messages.
func (l *logger) WithFields(args ...interface{}) Logger {
	defer recoverLogErr()
	f := fields(args...)
	base := l.base.WithFields(f)
	return &logger{l.logrus, base}
}

// PrintSimpleError prints out an error message with a red "ERROR:" prefix.
func PrintSimpleError(err error) {
	e := "Error:"
	if isColorTerminal(os.Stderr) {
		e = aurora.Red(e).String()
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", e, err.Error())
}

// recoverLogErr is used to recover from any panics during logging.
// Panics aren't expected of course, but logging should never crash
// a program, so this failsafe tries to prevent those crashes.
func recoverLogErr() {
	if r := recover(); r != nil {
		fmt.Println("Recovered from logging panic", r)
	}
}

// converts an argument list to a map, e.g.
// ("key", value, "key2", value2) => {"key": value, "key2", value2}
//
// Some arguments have special processing rules:
// - errors will be automatically expanded to have the key "error".
//
// - contexts will be searched for particular values, such as task ID.
//   The context values searched for are:
//   - taskID
//   - workerID
//
//   e.g.
//   ctx = context.WithValue("taskID", 1234)
//   fields("foo", fooval, ctx, err) returns
//   {"foo": fooval, "taskID", 1234, "error", err}.

func fields(args ...interface{}) map[string]interface{} {
	var expanded []interface{}
	for _, a := range args {
		switch x := a.(type) {
		case error:
			expanded = append(expanded, "error", a)
		case context.Context:
			if id, ok := x.Value("taskID").(string); ok {
				expanded = append(expanded, "taskID", id)
			}
			if id, ok := x.Value("workerID").(string); ok {
				expanded = append(expanded, "workerID", id)
			}
		default:
			expanded = append(expanded, a)
		}
	}

	f := make(map[string]interface{}, len(expanded)/2)
	if len(expanded) == 1 {
		f["unknown"] = expanded[0]
		return f
	}
	for i := 0; i < len(expanded); i += 2 {
		k := expanded[i].(string)
		v := expanded[i+1]
		f[k] = v
	}
	if len(expanded)%2 != 0 {
		f["unknown"] = expanded[len(expanded)-1]
	}
	return f
}
