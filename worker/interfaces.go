package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

type RunnerFactory func(c config.Worker, taskID string) Runner

type Runner interface {
	Run(context.Context)
}

type TaskService interface {
	TaskLogger

	Task() (*tes.Task, error)
	State() tes.State
	SetState(tes.State) error
}

// TaskLogger provides write access to a task's logs.
type TaskLogger interface {
	StartTime(t time.Time)
	EndTime(t time.Time)
	Outputs(o []*tes.OutputFileLog)
	Metadata(m map[string]string)

	ExecutorExitCode(i int, code int)
	ExecutorPorts(i int, ports []*tes.Ports)
	ExecutorHostIP(i int, ip string)
	ExecutorStartTime(i int, t time.Time)
	ExecutorEndTime(i int, t time.Time)

	AppendExecutorStdout(i int, s string)
	AppendExecutorStderr(i int, s string)
}
