package worker

import (
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

// TODO document behavior of slow consumer of task log updates

// RPCTask provides access to writing task logs over gRPC to the funnel server.
type RPCTask struct {
	client *schedClient
	taskID string
}

// StartTime updates the task's start time log.
func (r *RPCTask) StartTime(t time.Time) {
	r.client.UpdateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// EndTime updates the task's end time log.
func (r *RPCTask) EndTime(t time.Time) {
	r.client.UpdateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// Outputs updates the task's output file log.
func (r *RPCTask) Outputs(f []*tes.OutputFileLog) {
	r.client.UpdateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Outputs: f,
		},
	})
}

// Metadata updates the task's metadata log.
func (r *RPCTask) Metadata(m map[string]string) {
	r.client.UpdateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Metadata: m,
		},
	})
}

// ExecutorStartTime updates an executor's start time log.
func (r *RPCTask) ExecutorStartTime(i int, t time.Time) {
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorEndTime updates an executor's end time log.
func (r *RPCTask) ExecutorEndTime(i int, t time.Time) {
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorExitCode updates an executor's exit code log.
func (r *RPCTask) ExecutorExitCode(i int, x int) {
	log.Debug("Exit")
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			ExitCode: int32(x),
		},
	})
}

// ExecutorPorts updates an executor's ports log.
func (r *RPCTask) ExecutorPorts(i int, ports []*tes.Ports) {
	log.Debug("PORT")
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Ports: ports,
		},
	})
}

// ExecutorHostIP updates an executor's host IP log.
func (r *RPCTask) ExecutorHostIP(i int, ip string) {
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			HostIp: ip,
		},
	})
}

// AppendExecutorStdout appends to an executor's stdout log.
func (r *RPCTask) AppendExecutorStdout(i int, s string) {
	log.Debug("STDOUT")
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stdout: s,
		},
	})
}

// AppendExecutorStderr appends to an executor's stderr log.
func (r *RPCTask) AppendExecutorStderr(i int, s string) {
	log.Debug("STDERR")
	r.client.UpdateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stderr: s,
		},
	})
}
