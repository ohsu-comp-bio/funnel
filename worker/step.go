package worker

import (
	"context"
	"io"
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
	err := s.Event.StartTime(time.Now())
	if err != nil {
		return err
	}

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
			stdout, stderr = s.Event.StreamLogTail(subctx, s.Conf.LogTailSize, time.Duration(s.Conf.LogUpdateRate))
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
			err := s.Command.Stop()
			if err != nil {
				return err
			}
			<-done
			err = s.Event.EndTime(time.Now())
			if err != nil {
				return err
			}
			return ctx.Err()

		case result := <-done:
			err := s.Event.EndTime(time.Now())
			if err != nil {
				return err
			}
			err = s.Event.ExitCode(getExitCode(result))
			if err != nil {
				return err
			}
			return result
		}
	}
}
