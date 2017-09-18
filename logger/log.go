package logger

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/logrusorgru/aurora"
	"github.com/ohsu-comp-bio/funnel/util"
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
	f := util.ArgListToMap(args...)
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
	f := util.ArgListToMap(args...)
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
	f := util.ArgListToMap(args...)
	l.base.WithFields(f).Debug(msg)
}

// Info logs an info message
//
// After the first argument, arguments are key-value pairs which are written as structured logs.
//     log.Info("Some message here", "key1", value1, "key2", value2)
func (l *logger) Info(msg string, args ...interface{}) {
	defer recoverLogErr()
	f := util.ArgListToMap(args...)
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
	var f map[string]interface{}
	if len(args) == 1 {
		f = util.ArgListToMap("error", args[0])
	} else {
		f = util.ArgListToMap(args...)
	}
	l.base.WithFields(f).Error(msg)
}

// WithFields returns a new Logger instance with the given fields added to all log messages.
func (l *logger) WithFields(args ...interface{}) Logger {
	defer recoverLogErr()
	f := util.ArgListToMap(args...)
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
