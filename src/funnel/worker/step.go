package worker

import (
	"context"
	"funnel/config"
	"funnel/logger"
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"io"
	"time"
)

type stepRunner struct {
	TaskID  string
	Conf    config.Worker
	Num     int
	Cmd     *DockerCmd
	Log     logger.Logger
	Updates logUpdateChan
	IP      string
}

func (s *stepRunner) Run(ctx context.Context) error {
	log.Debug("Running step", "taskID", s.TaskID, "stepNum", s.Num)

	// Send update for host IP address.
	s.update(&tes.ExecutorLog{
		StartTime: time.Now().Format(time.RFC3339),
		HostIp:    s.IP,
	})

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the task is canceled.
	subctx, cleanup := context.WithCancel(ctx)
	defer cleanup()

	// tailLogs modifies the cmd Stdout/err fields, so should be called before Run.
	done := make(chan error, 1)

	stdout, stderr := s.logTails()
	defer stdout.Flush()
	defer stderr.Flush()

	ticker := time.NewTicker(s.Conf.LogUpdateRate)
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
			s.update(&tes.ExecutorLog{
				EndTime: time.Now().Format(time.RFC3339),
			})
			return ctx.Err()

		case <-ticker.C:
			stdout.Flush()
			stderr.Flush()

		case result := <-done:
			s.update(&tes.ExecutorLog{
				EndTime:  time.Now().Format(time.RFC3339),
				ExitCode: getExitCode(result),
			})
			return result
		}
	}
}

func (s *stepRunner) logTails() (*tailer, *tailer) {
	stdout, _ := newTailer(s.Conf.LogTailSize, func(c string) {
		s.update(&tes.ExecutorLog{Stdout: c})
	})
	stderr, _ := newTailer(s.Conf.LogTailSize, func(c string) {
		s.update(&tes.ExecutorLog{Stderr: c})
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
func (s *stepRunner) inspectContainer(ctx context.Context) {
	ports, err := s.Cmd.Inspect(ctx)
	if err != nil {
		s.Log.Error("Error inspecting container", err)
		return
	}
	s.update(&tes.ExecutorLog{
		Ports: ports,
	})
}

// update sends an update of the ExecutorLog of the currently running step.
// Used to update stdout/err logs, port mapping, etc.
func (s *stepRunner) update(log *tes.ExecutorLog) {
	s.Updates <- &pbf.UpdateExecutorLogsRequest{
		Id:   s.TaskID,
		Step: int64(s.Num),
		Log:  log,
	}
}
