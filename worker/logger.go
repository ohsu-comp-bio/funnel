package worker

import (
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io"
	"time"
)

// EventLogger wraps an events.Writer with helper functions.
type EventLogger struct {
	*events.AttemptWriter
}

// NewEventLogger returns a new EventLogger which writes events
// for a specific task + attempt to the given events.Writer.
func NewEventLogger(id string, attempt uint32, w events.Writer) EventLogger {
	return EventLogger{events.NewAttemptWriter(id, attempt, w)}
}

// Debug writes a SystemLog event with level "debug"
func (e EventLogger) Debug(msg string, fields map[string]string) error {
	return e.AttemptWriter.SystemLog(msg, "debug", fields)
}

// Info writes a SystemLog event with level "info"
func (e EventLogger) Info(msg string, fields map[string]string) error {
	return e.AttemptWriter.SystemLog(msg, "info", fields)
}

// Error writes a SystemLog event with level "error"
func (e EventLogger) Error(msg string, fields map[string]string) error {
	return e.AttemptWriter.SystemLog(msg, "error", fields)
}

// ExecutorStartTime writes an executor start time event.
func (e EventLogger) ExecutorStartTime(i int, t time.Time) error {
	return e.AttemptWriter.ExecutorStartTime(uint32(i), t)
}

// ExecutorEndTime writes an executor end time event.
func (e EventLogger) ExecutorEndTime(i int, t time.Time) error {
	return e.AttemptWriter.ExecutorEndTime(uint32(i), t)
}

// ExitCode writes an executor exit code event.
func (e EventLogger) ExitCode(i int, code int) error {
	return e.AttemptWriter.ExitCode(uint32(i), int32(code))
}

// Ports writes an executor ports event.
func (e EventLogger) Ports(i int, ports []*tes.Ports) error {
	return e.AttemptWriter.Ports(uint32(i), ports)
}

// HostIP writes an executor host IP event.
func (e EventLogger) HostIP(i int, ip string) error {
	return e.AttemptWriter.HostIP(uint32(i), ip)
}

// Stdout returns an io.Writer which writes an executor stdout event
// for each call to io.Writer.Write.
func (e EventLogger) Stdout(i int) io.Writer {
	return &events.StdoutWriter{e.AttemptWriter, uint32(i)}
}

// Stderr returns an io.Writer which writes an executor stderr event
// for each call to io.Writer.Write.
func (e EventLogger) Stderr(i int) io.Writer {
	return &events.StderrWriter{e.AttemptWriter, uint32(i)}
}
