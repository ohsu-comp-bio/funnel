package worker

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// NewLogTaskSvc returns a task service which writes task logs
// to the given logger instance (level info).
func NewLogTaskSvc(t *tes.Task, l logger.Logger) TaskService {
	return &logTaskService{l, t}
}

type logTaskService struct {
	log  logger.Logger
	task *tes.Task
}

func (ts *logTaskService) Task() (*tes.Task, error) {
	return ts.task, nil
}

func (ts *logTaskService) State() tes.State {
	return ts.task.State
}

func (ts *logTaskService) SetState(s tes.State) error {
	ts.log.Info("SetState", "State", s)
	ts.task.State = s
	return nil
}

func (ts *logTaskService) StartTime(t time.Time) {
	ts.log.Info("StartTime", "StartTime", t)
}

func (ts *logTaskService) EndTime(t time.Time) {
	ts.log.Info("EndTime", "EndTime", t)
}

func (ts *logTaskService) Outputs(o []*tes.OutputFileLog) {
	ts.log.Info("Outputs", "Outputs", o)
}

func (ts *logTaskService) Metadata(m map[string]string) {
	ts.log.Info("Metadata", "Metadata", m)
}

func (ts *logTaskService) ExecutorExitCode(i int, code int) {
	ts.log.Info("ExecutorExitCode", "ExecutorIndex", i, "ExecutorExitCode", code)
}

func (ts *logTaskService) ExecutorPorts(i int, ports []*tes.Ports) {
	ts.log.Info("ExecutorPorts", "ExecutorIndex", i, "ExecutorPorts", ports)
}

func (ts *logTaskService) ExecutorHostIP(i int, ip string) {
	ts.log.Info("ExecutorHostIP", "ExecutorIndex", i, "ExecutorHostIP", ip)
}

func (ts *logTaskService) ExecutorStartTime(i int, t time.Time) {
	ts.log.Info("ExecutorStartTime", "ExecutorIndex", i, "ExecutorStartTime", t)
}

func (ts *logTaskService) ExecutorEndTime(i int, t time.Time) {
	ts.log.Info("ExecutorEndTime", "ExecutorIndex", i, "ExecutorEndTime", t)
}

func (ts *logTaskService) AppendExecutorStdout(i int, s string) {
	ts.log.Info("AppendExecutorStdout", "ExecutorIndex", i, "AppendExecutorStdout", s)
}

func (ts *logTaskService) AppendExecutorStderr(i int, s string) {
	ts.log.Info("AppendExecutorStderr", "ExecutorIndex", i, "AppendExecutorStderr", s)
}
