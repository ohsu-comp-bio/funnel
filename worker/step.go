package worker

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
)

type stepWorker struct {
	Conf    config.Worker
	Command ContainerEngine
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
	_, out, err := s.Command.GetIO()
	if out != nil && out != (*os.File)(nil) {
		stdout = io.MultiWriter(out, stdout)
	}
	if err != nil && err != (*os.File)(nil) {
		stderr = io.MultiWriter(err, stderr)
	}
	s.Command.SetIO(nil, stdout, stderr)
	_, out, err = s.Command.GetIO()

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
