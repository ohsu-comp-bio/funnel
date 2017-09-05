package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
	"time"
)

// TODO document behavior of slow consumer of task log updates

// rpcTaskReader provides a TaskReader implementation which reads info over gRPC
// from an internal Funnel API.
type rpcTaskReader struct {
	client *taskClient
	taskID string
}

// Task returns the task descriptor.
func (r *rpcTaskReader) Task() (*tes.Task, error) {
	return r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *rpcTaskReader) State() tes.State {
	t, _ := r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id: r.taskID,
	})
	return t.GetState()
}

// rpcTaskWriter provides a TaskWriter implementation which writes updates over gRPC
// to an internal Funnel API.
type rpcTaskWriter struct {
	client        *taskClient
	taskID        string
	updateTimeout time.Duration
	log           logger.Logger
}

// State sets the state of the task.
func (r *rpcTaskWriter) State(s tes.State) {
	r.client.UpdateTaskState(context.Background(), &pbf.UpdateTaskStateRequest{
		Id:    r.taskID,
		State: s,
	})
}

// StartTime updates the task's start time log.
func (r *rpcTaskWriter) StartTime(t time.Time) {
	r.updateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// EndTime updates the task's end time log.
func (r *rpcTaskWriter) EndTime(t time.Time) {
	r.updateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// Outputs updates the task's output file log.
func (r *rpcTaskWriter) Outputs(f []*tes.OutputFileLog) {
	r.updateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Outputs: f,
		},
	})
}

// Metadata updates the task's metadata log.
func (r *rpcTaskWriter) Metadata(m map[string]string) {
	r.updateTaskLogs(&pbf.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Metadata: m,
		},
	})
}

// ExecutorStartTime updates an executor's start time log.
func (r *rpcTaskWriter) ExecutorStartTime(i int, t time.Time) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorEndTime updates an executor's end time log.
func (r *rpcTaskWriter) ExecutorEndTime(i int, t time.Time) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorExitCode updates an executor's exit code log.
func (r *rpcTaskWriter) ExecutorExitCode(i int, x int) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			ExitCode: int32(x),
		},
	})
}

// ExecutorPorts updates an executor's ports log.
func (r *rpcTaskWriter) ExecutorPorts(i int, ports []*tes.Ports) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Ports: ports,
		},
	})
}

// ExecutorHostIP updates an executor's host IP log.
func (r *rpcTaskWriter) ExecutorHostIP(i int, ip string) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			HostIp: ip,
		},
	})
}

// AppendExecutorStdout appends to an executor's stdout log.
func (r *rpcTaskWriter) AppendExecutorStdout(i int, s string) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stdout: s,
		},
	})
}

// AppendExecutorStderr appends to an executor's stderr log.
func (r *rpcTaskWriter) AppendExecutorStderr(i int, s string) {
	r.updateExecutorLogs(&pbf.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stderr: s,
		},
	})
}

func (r *rpcTaskWriter) updateExecutorLogs(up *pbf.UpdateExecutorLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateExecutorLogs(ctx, up)
	if err != nil {
		r.log.Error("Couldn't update executor logs", err)
	}
	cleanup()
	return err
}

func (r *rpcTaskWriter) updateTaskLogs(up *pbf.UpdateTaskLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateTaskLogs(ctx, up)
	if err != nil {
		r.log.Error("Couldn't update task logs", err)
	}
	cleanup()
	return err
}

type taskClient struct {
	tes.TaskServiceClient
	pbf.SchedulerServiceClient
}

func newTaskClient(conf config.Worker) (*taskClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		conf.ServerAddress,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		util.PerRPCPassword(conf.ServerPassword),
	)
	if err != nil {
		return nil, err
	}
	t := tes.NewTaskServiceClient(conn)
	s := pbf.NewSchedulerServiceClient(conn)
	return &taskClient{t, s}, nil
}
