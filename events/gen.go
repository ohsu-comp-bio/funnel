package events

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// AttemptGenerator helps create events for a single task attempt.
type AttemptGenerator struct {
	TaskID  string
	Attempt uint32
}

// State creates a state change event.
func (g *AttemptGenerator) State(s tes.State) *Event {
	return NewState(g.TaskID, g.Attempt, s)
}

// StartTime creates a task start time event.
func (g *AttemptGenerator) StartTime(t time.Time) *Event {
	return NewStartTime(g.TaskID, g.Attempt, t)
}

// EndTime creates a task end time event.
func (g *AttemptGenerator) EndTime(t time.Time) *Event {
	return NewEndTime(g.TaskID, g.Attempt, t)
}

// Outputs creates a task output file log event.
func (g *AttemptGenerator) Outputs(f []*tes.OutputFileLog) *Event {
	return NewOutputs(g.TaskID, g.Attempt, f)
}

// Metadata creates a task metadata log event.
func (g *AttemptGenerator) Metadata(m map[string]string) *Event {
	return NewMetadata(g.TaskID, g.Attempt, m)
}

// ExecutorStartTime creates an executor start time event
func (g *AttemptGenerator) ExecutorStartTime(i uint32, t time.Time) *Event {
	return NewExecutorStartTime(g.TaskID, g.Attempt, i, t)
}

// ExecutorEndTime creates an executor end time event.
func (g *AttemptGenerator) ExecutorEndTime(i uint32, t time.Time) *Event {
	return NewExecutorEndTime(g.TaskID, g.Attempt, i, t)
}

// ExitCode creates an executor exit code event
func (g *AttemptGenerator) ExitCode(i uint32, x int32) *Event {
	return NewExitCode(g.TaskID, g.Attempt, i, x)
}

// Ports creates an executor port metadata event
func (g *AttemptGenerator) Ports(i uint32, ports []*tes.Ports) *Event {
	return NewPorts(g.TaskID, g.Attempt, i, ports)
}

// HostIP creates an executor host IP metadata event
func (g *AttemptGenerator) HostIP(i uint32, ip string) *Event {
	return NewHostIP(g.TaskID, g.Attempt, i, ip)
}

// Stdout creates an executor stdout chunk event
func (g *AttemptGenerator) Stdout(i uint32, s string) *Event {
	return NewStdout(g.TaskID, g.Attempt, i, s)
}

// Stderr creates an executor stderr chunk event
func (g *AttemptGenerator) Stderr(i uint32, s string) *Event {
	return NewStderr(g.TaskID, g.Attempt, i, s)
}

// SystemLog creates a system log event
func (g *AttemptGenerator) SystemLog(msg, lvl string, fields map[string]string) *Event {
	return NewSystemLog(g.TaskID, g.Attempt, msg, lvl, fields)
}
