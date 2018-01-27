package events

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// NewTaskCreated creates a state change event.
func NewTaskCreated(task *tes.Task) *Event {
	return &Event{
		Id:        task.Id,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_CREATED,
		Data: &Event_Task{
			Task: task,
		},
	}
}

// NewState creates a state change event.
func NewState(taskID string, s tes.State) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_STATE,
		Data: &Event_State{
			State: s,
		},
	}
}

// NewStartTime creates a task start time event.
func NewStartTime(taskID string, t time.Time) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_START_TIME,
		Data: &Event_StartTime{
			StartTime: t.Format(time.RFC3339Nano),
		},
	}
}

// NewEndTime creates a task end time event.
func NewEndTime(taskID string, t time.Time) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_END_TIME,
		Data: &Event_EndTime{
			EndTime: t.Format(time.RFC3339Nano),
		},
	}
}

// NewOutputs creates a task output file log event.
func NewOutputs(taskID string, f []*tes.OutputFileLog) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_OUTPUTS,
		Data: &Event_Outputs{
			Outputs: &Outputs{
				Value: f,
			},
		},
	}
}

// NewMetadata creates a task metadata log event.
func NewMetadata(taskID string, m map[string]string) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_METADATA,
		Data: &Event_Metadata{
			Metadata: &Metadata{
				Value: m,
			},
		},
	}
}

// NewExecutorStartTime creates an executor start time event
// for the executor at the given index.
func NewExecutorStartTime(taskID string, t time.Time) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_START_TIME,
		Data: &Event_StartTime{
			StartTime: t.Format(time.RFC3339Nano),
		},
	}
}

// NewExecutorEndTime creates an executor end time event.
// for the executor at the given index.
func NewExecutorEndTime(taskID string, t time.Time) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_END_TIME,
		Data: &Event_EndTime{
			EndTime: t.Format(time.RFC3339Nano),
		},
	}
}

// NewExitCode creates an executor exit code event
// for the executor at the given index.
func NewExitCode(taskID string, x int32) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_EXIT_CODE,
		Data: &Event_ExitCode{
			ExitCode: x,
		},
	}
}

// NewStdout creates an executor stdout chunk event
// for the executor at the given index.
func NewStdout(taskID string, s string) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_STDOUT,
		Data: &Event_Stdout{
			Stdout: s,
		},
	}
}

// NewStderr creates an executor stderr chunk event
// for the executor at the given index.
func NewStderr(taskID string, s string) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_STDERR,
		Data: &Event_Stderr{
			Stderr: s,
		},
	}
}

// NewSystemLog creates an system log event.
func NewSystemLog(taskID string, lvl string, msg string, fields map[string]string) *Event {
	return &Event{
		Id:        taskID,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Type:      Type_SYSTEM_LOG,
		Data: &Event_SystemLog{
			SystemLog: &SystemLog{
				Msg:    msg,
				Level:  lvl,
				Fields: fields,
			},
		},
	}
}
