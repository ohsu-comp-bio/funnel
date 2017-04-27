package logutils

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"os"
)

// Configure configures the logging level and output path.
func Configure(conf config.Config) {
	logger.SetLevel(conf.LogLevel)
	logger.DisableTimestamp(!conf.TimestampLogs)

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems? syslog?
	if conf.LogPath != "" {
		logFile, err := os.OpenFile(
			conf.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666,
		)
		if err != nil {
			logger.Error("Can't open log output file", "path", conf.LogPath)
		} else {
			logger.SetOutput(logFile)
		}
	}
}
