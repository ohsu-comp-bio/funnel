package events

import (
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// NewState creates a state change event.
func NewState(taskID string, attempt uint32, s tes.State) *Event {
	return &Event{
		Type:      Type_STATE,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		State:     s,
	}
}

// NewStartTime creates a task start time event.
func NewStartTime(taskID string, attempt uint32, t time.Time) *Event {
	return &Event{
		Type:      Type_START_TIME,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		StartTime: Timestamp(t),
	}
}

// NewEndTime creates a task end time event.
func NewEndTime(taskID string, attempt uint32, t time.Time) *Event {
	return &Event{
		Type:      Type_END_TIME,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		EndTime:   Timestamp(t),
	}
}

// NewOutputs creates a task output file log event.
func NewOutputs(taskID string, attempt uint32, f []*tes.OutputFileLog) *Event {
	return &Event{
		Type:      Type_OUTPUTS,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		Outputs:   f,
	}
}

// NewMetadata creates a task metadata log event.
func NewMetadata(taskID string, attempt uint32, m map[string]string) *Event {
	return &Event{
		Type:      Type_METADATA,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		Metadata:  m,
	}
}

// NewExecutorStartTime creates an executor start time event
// for the executor at the given index.
func NewExecutorStartTime(taskID string, attempt uint32, index uint32, t time.Time) *Event {
	return &Event{
		Type:              Type_EXECUTOR_START_TIME,
		Id:                taskID,
		Attempt:           attempt,
		Timestamp:         ptypes.TimestampNow(),
		ExecutorStartTime: Timestamp(t),
		Index:             index,
	}
}

// NewExecutorEndTime creates an executor end time event.
// for the executor at the given index.
func NewExecutorEndTime(taskID string, attempt uint32, index uint32, t time.Time) *Event {
	return &Event{
		Type:            Type_EXECUTOR_END_TIME,
		Id:              taskID,
		Attempt:         attempt,
		Timestamp:       ptypes.TimestampNow(),
		ExecutorEndTime: Timestamp(t),
		Index:           index,
	}
}

// NewExitCode creates an executor exit code event
// for the executor at the given index.
func NewExitCode(taskID string, attempt uint32, index uint32, x int32) *Event {
	return &Event{
		Type:      Type_EXIT_CODE,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		ExitCode:  x,
		Index:     index,
	}
}

// NewPorts creates an executor port metadata event
// for the executor at the given index.
func NewPorts(taskID string, attempt uint32, index uint32, ports []*tes.Ports) *Event {
	return &Event{
		Type:      Type_PORTS,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		Ports:     ports,
		Index:     index,
	}
}

// NewHostIP creates an executor host IP metadata event
// for the executor at the given index.
func NewHostIP(taskID string, attempt uint32, index uint32, ip string) *Event {
	return &Event{
		Type:      Type_HOST_IP,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		HostIp:    ip,
		Index:     index,
	}
}

// NewStdout creates an executor stdout chunk event
// for the executor at the given index.
func NewStdout(taskID string, attempt uint32, index uint32, s string) *Event {
	return &Event{
		Type:      Type_STDOUT,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		Stdout:    s,
		Index:     index,
	}
}

// NewStderr creates an executor stderr chunk event
// for the executor at the given index.
func NewStderr(taskID string, attempt uint32, index uint32, s string) *Event {
	return &Event{
		Type:      Type_STDERR,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		Stderr:    s,
		Index:     index,
	}
}

// NewSystemLog creates an system log event.
func NewSystemLog(taskID string, attempt uint32, msg, lvl string, f map[string]string) *Event {
	return &Event{
		Type:      Type_SYSLOG,
		Id:        taskID,
		Attempt:   attempt,
		Timestamp: ptypes.TimestampNow(),
		SystemLog: &SystemLog{
			Msg:    msg,
			Level:  lvl,
			Fields: f,
		},
	}
}

// Timestamp converts a time.Time to a timestamp.
func Timestamp(t time.Time) *tspb.Timestamp {
	p, _ := ptypes.TimestampProto(t)
	return p
}

// TimestampString converts a timestamp to an RFC3339 formatted string.
func TimestampString(t *tspb.Timestamp) string {
	return ptypes.TimestampString(t)
}
