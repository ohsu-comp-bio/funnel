package worker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/ohsu-comp-bio/funnel/config"
	"io"
	"time"
)

type stepWorker struct {
	TaskID string
	Conf   config.Worker
	Num    int
	Cmd    *DockerCmd
	Log    EventLogger
	IP     string
}

func (s *stepWorker) Run(ctx context.Context) error {

	// Send update for host IP address.
	s.Log.ExecutorStartTime(s.Num, time.Now())
	s.Log.HostIP(s.Num, s.IP)

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the task is canceled.
	subctx, cleanup := context.WithCancel(ctx)
	defer cleanup()

	// tailLogs modifies the cmd Stdout/err fields, so should be called before Run.
	done := make(chan error, 1)

	s.logTails()

	go func() {
		done <- s.Cmd.Run()
	}()
	go s.inspectContainer(subctx)

	for {
		select {
		case <-ctx.Done():
			// Likely the task was canceled.
			s.Cmd.Stop()
			s.Log.ExecutorEndTime(s.Num, time.Now())
			return ctx.Err()

		case result := <-done:
			s.Log.ExecutorEndTime(s.Num, time.Now())
			s.Log.ExitCode(s.Num, getExitCode(result))
			return result
		}
	}
}

func (s *stepWorker) logTails() {
	if s.Cmd.Stdout != nil {
		s.Cmd.Stdout = io.MultiWriter(s.Cmd.Stdout, s.Log.Stdout(s.Num))
	}
	if s.Cmd.Stderr != nil {
		s.Cmd.Stderr = io.MultiWriter(s.Cmd.Stderr, s.Log.Stderr(s.Num))
	}
}

// inspectContainer calls Inspect on the DockerCmd, and sends an update with the results.
func (s *stepWorker) inspectContainer(ctx context.Context) {
	t := time.NewTimer(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.Log.Info("Inspecting container", nil)

			ports, err := s.Cmd.Inspect(ctx)
			if err != nil && !client.IsErrContainerNotFound(err) {
				s.Log.Error("Error inspecting container", map[string]string{
					"error": err.Error(),
				})
				break
			}
			s.Log.Ports(s.Num, ports)
			return
		}
	}
}
