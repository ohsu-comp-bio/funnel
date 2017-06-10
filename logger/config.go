package logger

import (
	"github.com/Sirupsen/logrus"
	"os"
)

// JSONFormatConfig provides configuration for the JSON logger format.
type JSONFormatConfig struct {
	DisableTimestamp bool
	TimestampFormat  string
}

// TextFormatConfig provides configuration for the text logger format.
type TextFormatConfig struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool
}

// Config provides configuration for a logger.
type Config struct {
	Level      string
	Formatter  string
	OutputFile string
	JSONFormat JSONFormatConfig
	TextFormat TextFormatConfig
}

// DefaultConfig returns a Config instance with default values.
func DefaultConfig() Config {
	return Config{
		Level:     "info",
		Formatter: "text",
		TextFormat: TextFormatConfig{
			TimestampFormat: logrus.DefaultTimestampFormat,
		},
	}
}

// DebugConfig returns a Config instance with default values useful for testing/debugging.
func DebugConfig() Config {
	return Config{
		Level:     "debug",
		Formatter: "text",
		TextFormat: TextFormatConfig{
			ForceColors:     true,
			TimestampFormat: logrus.DefaultTimestampFormat,
		},
	}
}

// Configure configures the logging level and output path.
func (l *logger) Configure(conf Config) {
	l.SetLevel(conf.Level)

	switch conf.Formatter {
	case "json":
		l.SetFormatter(&jsonFormatter{
			conf: conf.JSONFormat,
		})

	case "":
		// Default to text
		fallthrough
	case "text":
		fallthrough
	default:
		// Default to text
		l.SetFormatter(&textFormatter{
			conf.TextFormat,
			jsonFormatter{
				conf: conf.JSONFormat,
			},
		})
	}

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems? syslog?
	if conf.OutputFile != "" {
		logFile, err := os.OpenFile(
			conf.OutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666,
		)
		if err != nil {
			l.Error("Can't open log output", "output", conf.OutputFile)
		} else {
			l.SetOutput(logFile)
		}
	}
}
