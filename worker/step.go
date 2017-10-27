package worker

import (
	"context"
	"github.com/docker/docker/client"
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

	// tailLogs modifies the cmd Stdout/err fields, so should be called before Run.
	done := make(chan error, 1)

	stdout, stderr := s.logTails()
	defer stdout.Flush()
	defer stderr.Flush()

	ticker := time.NewTicker(s.Conf.UpdateRate)
	defer ticker.Stop()

	go func() {
		done <- s.Cmd.Run()
	}()
	go s.inspectContainer(subctx)

	for {
		select {
		case <-ctx.Done():
			// Likely the task was canceled.
			s.Cmd.Stop()
			s.Event.EndTime(time.Now())
			return ctx.Err()

		case <-ticker.C:
			stdout.Flush()
			stderr.Flush()

		case result := <-done:
			s.Event.EndTime(time.Now())
			return result
		}
	}
}

func (s *stepWorker) logTails() (*tailer, *tailer) {
	stdout, _ := newTailer(s.Conf.BufferSize, func(c string) {
		s.Event.Stdout(c)
	})
	stderr, _ := newTailer(s.Conf.BufferSize, func(c string) {
		s.Event.Stderr(c)
	})
	if s.Cmd.Stdout != nil {
		s.Cmd.Stdout = io.MultiWriter(s.Cmd.Stdout, stdout)
	}
	if s.Cmd.Stderr != nil {
		s.Cmd.Stderr = io.MultiWriter(s.Cmd.Stderr, stderr)
	}
	return stdout, stderr
}

// inspectContainer calls Inspect on the DockerCmd, and sends an update with the results.
func (s *stepWorker) inspectContainer(ctx context.Context) {
	t := time.NewTimer(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.Event.Info("Inspecting container")

			ports, err := s.Cmd.Inspect(ctx)
			if err != nil && !client.IsErrContainerNotFound(err) {
				s.Event.Error("Error inspecting container", err)
				break
			}
			s.Event.Ports(ports)
			return
		}
	}
}
