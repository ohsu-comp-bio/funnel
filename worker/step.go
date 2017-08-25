package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"io"
	"time"
)

type stepWorker struct {
	TaskID     string
	Conf       config.Worker
	Num        int
	Cmd        *DockerCmd
	Log        logger.Logger
	TaskLogger TaskLogger
	IP         string
}

func (s *stepWorker) Run(ctx context.Context) error {
	log.Debug("Running step", "taskID", s.TaskID, "stepNum", s.Num)

	// Send update for host IP address.
	s.TaskLogger.ExecutorStartTime(s.Num, time.Now())
	s.TaskLogger.ExecutorHostIP(s.Num, s.IP)

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
			s.TaskLogger.ExecutorEndTime(s.Num, time.Now())
			return ctx.Err()

		case <-ticker.C:
			stdout.Flush()
			stderr.Flush()

		case result := <-done:
			s.TaskLogger.ExecutorEndTime(s.Num, time.Now())
			s.TaskLogger.ExecutorExitCode(s.Num, getExitCode(result))
			return result
		}
	}
}

func (s *stepWorker) logTails() (*tailer, *tailer) {
	stdout, _ := newTailer(s.Conf.BufferSize, func(c string) {
		s.TaskLogger.AppendExecutorStdout(s.Num, c)
	})
	stderr, _ := newTailer(s.Conf.BufferSize, func(c string) {
		s.TaskLogger.AppendExecutorStderr(s.Num, c)
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
	ports, err := s.Cmd.Inspect(ctx)
	if err != nil {
		s.Log.Error("Error inspecting container", err)
		return
	}
	s.TaskLogger.ExecutorPorts(s.Num, ports)
}
