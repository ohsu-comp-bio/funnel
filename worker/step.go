package worker

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
)

type stepWorker struct {
	Conf    config.Worker
	Command *DockerCommand
	Event   *events.ExecutorWriter
	IP      string
}

func (s *stepWorker) Run(ctx context.Context) error {
	// WaitGroup to block closing the worker until the last event is written from events/executor.go:StreamLogTail()
	var wg sync.WaitGroup
	defer wg.Wait()
	s.Event.StartTime(time.Now())

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the task is canceled.
	subctx, cleanup := context.WithCancel(context.Background())
	defer cleanup()

	done := make(chan error, 1)
	var stdout io.Writer
	var stderr io.Writer

	// Tail the stdout/err log streams.
	if s.Conf.LogTailSize > 0 {
		if s.Conf.LogUpdateRate > 0 {
			stdout, stderr = s.Event.StreamLogTail(subctx, s.Conf.LogTailSize, time.Duration(s.Conf.LogUpdateRate), &wg)
		} else {
			stdout, stderr = s.Event.LogTail(subctx, s.Conf.LogTailSize)
		}
	}

	// Capture stdout/err to file.
	if s.Command.Stdout != nil {
		stdout = io.MultiWriter(s.Command.Stdout, stdout)
	}
	if s.Command.Stderr != nil {
		stderr = io.MultiWriter(s.Command.Stderr, stderr)
	}
	s.Command.Stdout = stdout
	s.Command.Stderr = stderr

	go func() {
		done <- s.Command.Run(subctx)
	}()

	for {
		select {
		case <-ctx.Done():
			// Likely the task was canceled.
			s.Command.Stop()
			<-done
			s.Event.EndTime(time.Now())
			return ctx.Err()

		case result := <-done:
			s.Event.EndTime(time.Now())
			s.Event.ExitCode(getExitCode(result))
			return result
		}
	}
}
