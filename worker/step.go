package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"io"
	"time"
)

type stepWorker struct {
	Conf    config.Worker
	Command *DockerCommand
	Event   *events.ExecutorWriter
	IP      string
}

func (s *stepWorker) Run(ctx context.Context) error {
	s.Event.StartTime(time.Now())

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the task is canceled.
	subctx, cleanup := context.WithCancel(ctx)
	defer cleanup()

	// Tail the stdout/err log streams.
	stdout, stderr := s.Event.TailLogs(subctx, s.Conf.BufferSize, s.Conf.UpdateRate)
	if s.Command.Stdout != nil {
		stdout = io.MultiWriter(s.Command.Stdout, stdout)
	}
	if s.Command.Stderr != nil {
		stderr = io.MultiWriter(s.Command.Stderr, stderr)
	}
	s.Command.Stdout = stdout
	s.Command.Stderr = stderr

	done := make(chan error, 1)
	go func() {
		done <- s.Command.Run()
	}()

	for {
		select {
		case <-ctx.Done():
			// Likely the task was canceled.
			s.Command.Stop()
			s.Event.EndTime(time.Now())
			return ctx.Err()

		case result := <-done:
			s.Event.EndTime(time.Now())
			s.Event.ExitCode(getExitCode(result))
			return result
		}
	}
}
