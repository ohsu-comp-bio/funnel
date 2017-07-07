package worker

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// NewThinTaskLogger returns a task logger which writes task logs
// to the given logger instance (level info).
func NewThinTaskLogger(l logger.Logger) TaskLogger {
	return &thinTaskLogger{l}
}

type thinTaskLogger struct {
	log logger.Logger
}

func (p *thinTaskLogger) StartTime(t time.Time) {
	p.log.Info("StartTime", "StartTime", t)
}

func (p *thinTaskLogger) EndTime(t time.Time) {
	p.log.Info("EndTime", "EndTime", t)
}

func (p *thinTaskLogger) Outputs(o []*tes.OutputFileLog) {
	p.log.Info("Outputs", "Outputs", o)
}

func (p *thinTaskLogger) Metadata(m map[string]string) {
	p.log.Info("Metadata", "Metadata", m)
}

func (p *thinTaskLogger) ExecutorExitCode(i int, code int) {
	p.log.Info("ExecutorExitCode", "ExecutorIndex", i, "ExecutorExitCode", code)
}

func (p *thinTaskLogger) ExecutorPorts(i int, ports []*tes.Ports) {
	p.log.Info("ExecutorPorts", "ExecutorIndex", i, "ExecutorPorts", ports)
}

func (p *thinTaskLogger) ExecutorHostIP(i int, ip string) {
	p.log.Info("ExecutorHostIP", "ExecutorIndex", i, "ExecutorHostIP", ip)
}

func (p *thinTaskLogger) ExecutorStartTime(i int, t time.Time) {
	p.log.Info("ExecutorStartTime", "ExecutorIndex", i, "ExecutorStartTime", t)
}

func (p *thinTaskLogger) ExecutorEndTime(i int, t time.Time) {
	p.log.Info("ExecutorEndTime", "ExecutorIndex", i, "ExecutorEndTime", t)
}

func (p *thinTaskLogger) AppendExecutorStdout(i int, s string) {
	p.log.Info("AppendExecutorStdout", "ExecutorIndex", i, "AppendExecutorStdout", s)
}

func (p *thinTaskLogger) AppendExecutorStderr(i int, s string) {
	p.log.Info("AppendExecutorStderr", "ExecutorIndex", i, "AppendExecutorStderr", s)
}
