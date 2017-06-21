package logger

import (
	"io"
)

var global = New("funnel")

// SetLevel sets the logging level for the global logger.
func SetLevel(lvl string) {
	global.SetLevel(lvl)
}

// SetFormatter sets the formatter for the global logger
func SetFormatter(f Formatter) {
	global.SetFormatter(f)
}

// SetOutput sets the output for the global logger
func SetOutput(w io.Writer) {
	global.SetOutput(w)
}

// Discard discards the output for the global logger
func Discard() {
	global.Discard()
}

// Debug logs to the global logger at the Debug level
func Debug(msg string, args ...interface{}) {
	global.Debug(msg, args...)
}

// Info logs to the global logger at the Info level
func Info(msg string, args ...interface{}) {
	global.Info(msg, args...)
}

// Error logs to the global logger at the Error level
func Error(msg string, args ...interface{}) {
	global.Error(msg, args...)
}

// WithFields returns a child logger of the global logger with the given fields.
func WithFields(args ...interface{}) Logger {
	return global.WithFields(args...)
}

// Configure configures the global logger.
func Configure(c Config) {
	global.Configure(c)
}

// NewSubLogger returns a new sub-logger instance.
func NewSubLogger(ns string, args ...interface{}) Logger {
	return global.NewSubLogger(ns, args...)
}
