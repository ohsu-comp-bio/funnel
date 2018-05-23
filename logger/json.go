package logger

import (
	"github.com/sirupsen/logrus"
)

type jsonFormatter struct {
	conf JSONFormatConfig
	fmt  *logrus.JSONFormatter
}

func (f *jsonFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.fmt == nil {
		f.fmt = &logrus.JSONFormatter{
			DisableTimestamp: f.conf.DisableTimestamp,
			TimestampFormat:  f.conf.TimestampFormat,
		}
	}
	return f.fmt.Format(entry)
}
