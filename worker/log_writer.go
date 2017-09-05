package worker

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

type logTaskWriter struct {
	logger.Logger
}

func (l *logTaskWriter) State(s tes.State) {
	l.Info("State", "State", s.String())
}

func (l *logTaskWriter) StartTime(t time.Time) {
	l.Info("StartTime", "StartTime", t)
}

func (l *logTaskWriter) EndTime(t time.Time) {
	l.Info("EndTime", "EndTime", t)
}

func (l *logTaskWriter) Outputs(o []*tes.OutputFileLog) {
	l.Info("Outputs", "Outputs", o)
}

func (l *logTaskWriter) Metadata(m map[string]string) {
	l.Info("Metadata", "Metadata", m)
}

func (l *logTaskWriter) ExecutorExitCode(i int, code int) {
	l.Info("ExecutorExitCode", "ExecutorIndex", i, "ExecutorExitCode", code)
}

func (l *logTaskWriter) ExecutorPorts(i int, ports []*tes.Ports) {
	l.Info("ExecutorPorts", "ExecutorIndex", i, "ExecutorPorts", ports)
}

func (l *logTaskWriter) ExecutorHostIP(i int, ip string) {
	l.Info("ExecutorHostIP", "ExecutorIndex", i, "ExecutorHostIP", ip)
}

func (l *logTaskWriter) ExecutorStartTime(i int, t time.Time) {
	l.Info("ExecutorStartTime", "ExecutorIndex", i, "ExecutorStartTime", t)
}

func (l *logTaskWriter) ExecutorEndTime(i int, t time.Time) {
	l.Info("ExecutorEndTime", "ExecutorIndex", i, "ExecutorEndTime", t)
}

func (l *logTaskWriter) AppendExecutorStdout(i int, s string) {
	l.Info("AppendExecutorStdout", "ExecutorIndex", i, "AppendExecutorStdout", s)
}

func (l *logTaskWriter) AppendExecutorStderr(i int, s string) {
	l.Info("AppendExecutorStderr", "ExecutorIndex", i, "AppendExecutorStderr", s)
}
