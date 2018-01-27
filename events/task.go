package events

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/ring"
  "golang.org/x/time/rate"
  "io"
	"time"
)

var bg = context.Background()

// TaskWriter is a type that generates and writes task events.
type TaskWriter struct {
  id string
	out Writer
}

// NewTaskWriter returns a TaskWriter instance.
func NewTaskWriter(taskID string, w Writer) *TaskWriter {
	return &TaskWriter{
    id: taskID,
		out: w,
	}
}

// State sets the state of the task.
func (tw *TaskWriter) State(s tes.State) error {
	return tw.out.WriteEvent(bg, NewState(tw.id, s))
}

// StartTime updates the task's start time log.
func (tw *TaskWriter) StartTime(t time.Time) error {
	return tw.out.WriteEvent(bg, NewStartTime(tw.id, t))
}

// EndTime updates the task's end time log.
func (tw *TaskWriter) EndTime(t time.Time) error {
	return tw.out.WriteEvent(bg, NewEndTime(tw.id, t))
}

// Outputs updates the task's output file log.
func (tw *TaskWriter) Outputs(f []*tes.OutputFileLog) error {
	return tw.out.WriteEvent(bg, NewOutputs(tw.id, f))
}

// Metadata updates the task's metadata log.
func (tw *TaskWriter) Metadata(m map[string]string) error {
	return tw.out.WriteEvent(bg, NewMetadata(tw.id, m))
}

// Info creates an info level system log message.
func (tw *TaskWriter) Info(msg string, args ...interface{}) error {
	return tw.out.WriteEvent(bg, NewSystemLog(tw.id, "info", msg, fields(args...)))
}

// Debug creates a debug level system log message.
func (tw *TaskWriter) Debug(msg string, args ...interface{}) error {
	return tw.out.WriteEvent(bg, NewSystemLog(tw.id, "debug", msg, fields(args...)))
}

// Error creates an error level system log message.
func (tw *TaskWriter) Error(msg string, args ...interface{}) error {
	return tw.out.WriteEvent(bg, NewSystemLog(tw.id, "error", msg, fields(args...)))
}

// Warn creates a warning level system log message.
func (tw *TaskWriter) Warn(msg string, args ...interface{}) error {
	return tw.out.WriteEvent(bg, NewSystemLog(tw.id, "warn", msg, fields(args...)))
}

// Stdout appends to an executor's stdout log.
func (tw *TaskWriter) Stdout(s string) error {
	return tw.out.WriteEvent(bg, NewStdout(tw.id, s))
}

// Stderr appends to an executor's stderr log.
func (tw *TaskWriter) Stderr(s string) error {
	return tw.out.WriteEvent(bg, NewStderr(tw.id, s))
}

// ExitCode updates an executor's exit code log.
func (tw *TaskWriter) ExitCode(x int) error {
	return tw.out.WriteEvent(bg, NewExitCode(tw.id, int32(x)))
}

// TailLogs returns stdout/err io.Writers which will track the
// tail of the content (up to "size") and emit events. Events
// are rate limited by "interval", e.g. a max of one event every
// 5 seconds.
func (tw *TaskWriter) TailLogs(ctx context.Context, size int64, interval time.Duration) (stdout, stderr io.Writer) {
	return TailLogs(ctx, tw.id, size, interval, tw.out)
}

// TailLogs returns stdout/err io.Writers which will track the
// tail of the content (up to "size") and emit events. Events
// are rate limited by "interval", e.g. a max of one event every
// 5 seconds.
func TailLogs(ctx context.Context, taskID string, size int64, interval time.Duration, out Writer) (stdout, stderr io.Writer) {

	// The rate limiter allows the input writers to trigger events
	// immediately, without waiting for the ticker, as long as
	// they are not exceeding the rate limit.
	limiter := rate.NewLimiter(rate.Every(interval), 1)

	stdoutbuf := ring.NewBuffer(size)
	stderrbuf := ring.NewBuffer(size)
	stdoutch := make(chan []byte)
	stderrch := make(chan []byte)
	eventch := make(chan *Event)
	// Used as an immediate timeout for flush()
	immediate := make(chan time.Time)
	close(immediate)

	flush := func(buf *ring.Buffer, t Type, timeout <-chan time.Time) {
		// Only flush if new bytes have been written to the buffer.
		if buf.TotalWritten() == 0 {
			return
		}

		// Create the event
		var e *Event
		s := buf.String()
		switch t {
		case Type_STDOUT:
			e = NewStdout(taskID, s)
		case Type_STDERR:
			e = NewStderr(taskID, s)
		}

		// Send the event to the routine which is writing out events.
		// If it's busy, don't wait because it will block the stdout/err streams
		// writing into the logs. The logs will be flushed again soon anyway.
		select {
		case eventch <- e:
			// The writer routine accepted the event, so reset the buffer byte count.
			buf.ResetTotalWritten()
		case <-timeout:
			// The writer was busy, do nothing.
		}
	}

	flushboth := func(timeout <-chan time.Time) {
		flush(stdoutbuf, Type_STDOUT, timeout)
		flush(stderrbuf, Type_STDERR, timeout)
	}

	// There are two routines below, one for accepting input, one for writing
	// out events. They are separated so that writing out events does not block
	// the input writes. If input writes are faster than output event writes,
	// flush() calls will be dropped. This is ok, because we're flushing the
	// whole buffer (log tail) every tick, so when the output event writer
	// catches up, it will write the new, complete tail.

	// output event writer routine
	go func() {
		for e := range eventch {
			ctx, cancel := context.WithTimeout(bg, time.Second*5)
			out.WriteEvent(ctx, e)
			cancel()
		}
	}()

	// input writes and flush routine.
	go func() {
		// The ticker helps ensure content gets flushed at a regular
		// interval, so nothing is buffered for too long.
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				timeout := time.After(time.Second * 5)
				flushboth(timeout)
				close(eventch)
				return
			case <-ticker.C:
				w := stdoutbuf.TotalWritten() + stderrbuf.TotalWritten()
				// Don't use a limiter token if not content has been written.
				if w > 0 && limiter.Allow() {
					flushboth(immediate)
				}
			case b := <-stdoutch:
				stdoutbuf.Write(b)
				if limiter.Allow() {
					flushboth(immediate)
				}
			case b := <-stderrch:
				stderrbuf.Write(b)
				if limiter.Allow() {
					flushboth(immediate)
				}
			}
		}
	}()

	return &logTailWriter{stdoutch}, &logTailWriter{stderrch}
}

type logTailWriter struct {
	ch chan<- []byte
}

func (l *logTailWriter) Write(p []byte) (n int, err error) {
	l.ch <- p
	return len(p), nil
}
