package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"io"
	"time"
)

type stepWorker struct {
	Conf  config.Worker
	Cmd   *DockerCmd
	Event *events.ExecutorWriter
	IP    string
}

func (s *stepWorker) Run(ctx context.Context) error {
	// Send update for host IP address.
	s.Event.StartTime(time.Now())
	s.Event.HostIP(s.IP)

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the task is canceled.
	subctx, cleanup := context.WithCancel(ctx)
	defer cleanup()

	// Tail the stdout/err log streams.
	stdout, stderr := s.Event.TailLogs(subctx, s.Conf.BufferSize, s.Conf.UpdateRate)
	if s.Cmd.Stdout != nil {
		stdout = io.MultiWriter(s.Cmd.Stdout, stdout)
	}
	if s.Cmd.Stderr != nil {
		stderr = io.MultiWriter(s.Cmd.Stderr, stderr)
	}
	s.Cmd.Stdout = stdout
	s.Cmd.Stderr = stderr

	done := make(chan error, 1)
	go func() {
		done <- s.Cmd.Run()
	}()

	for {
		select {
		case <-ctx.Done():
			// Likely the task was canceled.
			s.Cmd.Stop()
			s.Event.EndTime(time.Now())
			return ctx.Err()

		case result := <-done:
			s.Event.EndTime(time.Now())
			s.Event.ExitCode(getExitCode(result))
			return result
		}
	}
}
