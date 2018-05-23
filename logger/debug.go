package logger

var debug = NewLogger("debug", DebugConfig())

// Debug logs debug messages to a global logger.
func Debug(msg string, args ...interface{}) {
	debug.Debug(msg, args...)
}

// DebugConfig returns a Config instance with default values useful for testing/debugging.
func DebugConfig() Config {
	return Config{
		Level:     "debug",
		Formatter: "text",
		TextFormat: TextFormatConfig{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: defaultTimestampFormat,
		},
	}
}
