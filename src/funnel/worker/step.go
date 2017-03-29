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
	JobID   string
	Conf    config.Worker
	Num     int
	Cmd     *DockerCmd
	Log     logger.Logger
	Updates logUpdateChan
	IP      string
}

func (s *stepRunner) Run(ctx context.Context) error {
	log.Debug("Running step", "jobID", s.JobID, "stepNum", s.Num)

	// Send update for host IP address.
	s.update(&tes.JobLog{
		HostIP: s.IP,
	})

	// subctx helps ensure that these goroutines are cleaned up,
	// even when the job is canceled.
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
			// Likely the job was canceled.
			s.Cmd.Stop()
			return ctx.Err()

		case <-ticker.C:
			stdout.Flush()
			stderr.Flush()

		case result := <-done:
			s.update(&tes.JobLog{
				ExitCode: getExitCode(result),
			})
			return result
		}
	}
}

func (s *stepRunner) logTails() (*tailer, *tailer) {
	stdout, _ := newTailer(s.Conf.LogTailSize, func(c string) {
		s.update(&tes.JobLog{Stdout: c})
	})
	stderr, _ := newTailer(s.Conf.LogTailSize, func(c string) {
		s.update(&tes.JobLog{Stderr: c})
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
	s.update(&tes.JobLog{
		Ports: ports,
	})
}

// update sends an update of the JobLog of the currently running step.
// Used to update stdout/err logs, port mapping, etc.
func (s *stepRunner) update(log *tes.JobLog) {
	s.Updates <- &pbf.UpdateJobLogsRequest{
		Id:   s.JobID,
		Step: int64(s.Num),
		Log:  log,
	}
}
