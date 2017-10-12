package events

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// TaskBuilder aggregates events into an in-memory Task object.
type TaskBuilder struct {
	*tes.Task
}

// Write updates the Task object.
func (tb TaskBuilder) Write(ev *Event) error {
	t := tb.Task
	t.Id = ev.Id
	attempt := int(ev.Attempt)
	index := int(ev.Index)

	switch ev.Type {
	case Type_TASK_STATE:
		to := ev.GetState()
		if err := tes.ValidateTransition(t.GetState(), ev.GetState()); err != nil {
			return err
		}
		t.State = to

	case Type_TASK_START_TIME:
		t.GetTaskLog(attempt).StartTime = TimestampString(ev.GetStartTime())

	case Type_TASK_END_TIME:
		t.GetTaskLog(attempt).EndTime = TimestampString(ev.GetEndTime())

	case Type_TASK_OUTPUTS:
		t.GetTaskLog(attempt).Outputs = ev.GetOutputs().Value

	case Type_TASK_METADATA:
		t.GetTaskLog(attempt).Metadata = ev.GetMetadata().Value

	case Type_EXECUTOR_START_TIME:
		t.GetExecLog(attempt, index).StartTime = TimestampString(ev.GetStartTime())

	case Type_EXECUTOR_END_TIME:
		t.GetExecLog(attempt, index).EndTime = TimestampString(ev.GetEndTime())

	case Type_EXECUTOR_EXIT_CODE:
		t.GetExecLog(attempt, index).ExitCode = ev.GetExitCode()

	case Type_EXECUTOR_HOST_IP:
		t.GetExecLog(attempt, index).HostIp = ev.GetHostIp()

	case Type_EXECUTOR_PORTS:
		t.GetExecLog(attempt, index).Ports = ev.GetPorts().Value

	case Type_EXECUTOR_STDOUT:
		t.GetExecLog(attempt, index).Stdout += ev.GetStdout()

	case Type_EXECUTOR_STDERR:
		t.GetExecLog(attempt, index).Stderr += ev.GetStderr()
	}

	return nil
}
