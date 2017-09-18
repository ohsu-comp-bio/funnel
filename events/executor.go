package events

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// ExecutorGenerator is a type that generates Events for an Executor
// of a Task
type ExecutorGenerator struct {
	taskID  string
	attempt uint32
	index   uint32
	sys     *SystemLogGenerator
}

// NewExecutorGenerator returns a ExecutorGenerator instance.
func NewExecutorGenerator(taskID string, attempt uint32, index uint32) *ExecutorGenerator {
	return &ExecutorGenerator{taskID, attempt, index, &SystemLogGenerator{taskID, attempt, index}}
}

// StartTime updates an executor's start time log.
func (eg *ExecutorGenerator) StartTime(t time.Time) *Event {
	return NewExecutorStartTime(eg.taskID, eg.attempt, eg.index, t)
}

// EndTime updates an executor's end time log.
func (eg *ExecutorGenerator) EndTime(t time.Time) *Event {
	return NewExecutorEndTime(eg.taskID, eg.attempt, eg.index, t)
}

// ExitCode updates an executor's exit code log.
func (eg *ExecutorGenerator) ExitCode(x int) *Event {
	return NewExitCode(eg.taskID, eg.attempt, eg.index, int32(x))
}

// Ports updates an executor's ports log.
func (eg *ExecutorGenerator) Ports(ports []*tes.Ports) *Event {
	return NewPorts(eg.taskID, eg.attempt, eg.index, ports)
}

// HostIP updates an executor's host IP log.
func (eg *ExecutorGenerator) HostIP(ip string) *Event {
	return NewHostIP(eg.taskID, eg.attempt, eg.index, ip)
}

// Stdout appends to an executor's stdout log.
func (eg *ExecutorGenerator) Stdout(s string) *Event {
	return NewStdout(eg.taskID, eg.attempt, eg.index, s)
}

// Stderr appends to an executor's stderr log.
func (eg *ExecutorGenerator) Stderr(s string) *Event {
	return NewStderr(eg.taskID, eg.attempt, eg.index, s)
}

// Info creates an info level system log message.
func (eg *ExecutorGenerator) Info(msg string, args ...interface{}) *Event {
	return eg.sys.Info(msg, args...)
}

// Debug creates a debug level system log message.
func (eg *ExecutorGenerator) Debug(msg string, args ...interface{}) *Event {
	return eg.sys.Debug(msg, args...)
}

// Error creates an error level system log message.
func (eg *ExecutorGenerator) Error(msg string, args ...interface{}) *Event {
	return eg.sys.Error(msg, args...)
}

// ExecutorWriter is a type that generates and writes executor events.
type ExecutorWriter struct {
	gen *ExecutorGenerator
	sys *SystemLogWriter
	out Writer
}

// NewExecutorWriter returns a ExecutorWriter instance.
func NewExecutorWriter(taskID string, attempt uint32, index uint32, logLevel string, w Writer) *ExecutorWriter {
	g := NewExecutorGenerator(taskID, attempt, index)
	return &ExecutorWriter{
		gen: g,
		out: w,
		sys: &SystemLogWriter{logLevel, g.sys, w},
	}
}

// StartTime updates the task's start time log.
func (ew *ExecutorWriter) StartTime(t time.Time) error {
	return ew.out.Write(ew.gen.StartTime(t))
}

// EndTime updates the task's end time log.
func (ew *ExecutorWriter) EndTime(t time.Time) error {
	return ew.out.Write(ew.gen.EndTime(t))
}

// ExitCode updates an executor's exit code log.
func (ew *ExecutorWriter) ExitCode(x int) error {
	return ew.out.Write(ew.gen.ExitCode(x))
}

// Ports updates an executor's ports log.
func (ew *ExecutorWriter) Ports(ports []*tes.Ports) error {
	return ew.out.Write(ew.gen.Ports(ports))
}

// HostIP updates an executor's host IP log.
func (ew *ExecutorWriter) HostIP(ip string) error {
	return ew.out.Write(ew.gen.HostIP(ip))
}

// Stdout appends to an executor's stdout log.
func (ew *ExecutorWriter) Stdout(s string) error {
	return ew.out.Write(ew.gen.Stdout(s))
}

// Stderr appends to an executor's stderr log.
func (ew *ExecutorWriter) Stderr(s string) error {
	return ew.out.Write(ew.gen.Stderr(s))
}

// Info writes an info level system log message.
func (ew *ExecutorWriter) Info(msg string, args ...interface{}) error {
	return ew.sys.Info(msg, args...)
}

// Debug writes a debug level system log message.
func (ew *ExecutorWriter) Debug(msg string, args ...interface{}) error {
	return ew.sys.Debug(msg, args...)
}

// Error writes an error level system log message.
func (ew *ExecutorWriter) Error(msg string, args ...interface{}) error {
	return ew.sys.Error(msg, args...)
}
