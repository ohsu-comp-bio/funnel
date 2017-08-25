package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	tl "github.com/ohsu-comp-bio/funnel/proto/tasklogger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"google.golang.org/grpc"
	"time"
)

// TODO document behavior of slow consumer of task log updates

// RPCTask provides access to writing task logs over gRPC to the funnel server.
type RPCTask struct {
	client        *taskClient
	taskID        string
	updateTimeout time.Duration
}

func newRPCTask(conf config.Worker, taskID string) (*RPCTask, error) {
	client, err := newTaskClient(conf)
	if err != nil {
		return nil, err
	}
	return &RPCTask{client, taskID, conf.UpdateTimeout}, nil
}

// Task returns the task descriptor.
func (r *RPCTask) Task() (*tes.Task, error) {
	return r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   r.taskID,
		View: tes.TaskView_FULL,
	})
}

// State returns the current state of the task.
func (r *RPCTask) State() tes.State {
	t, _ := r.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id: r.taskID,
	})
	return t.GetState()
}

// SetState sets the state of the task.
func (r *RPCTask) SetState(s tes.State) error {
	_, err := r.client.UpdateTaskState(context.Background(), &tl.UpdateTaskStateRequest{
		Id:    r.taskID,
		State: s,
	})
	return err
}

// StartTime updates the task's start time log.
func (r *RPCTask) StartTime(t time.Time) {
	r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// EndTime updates the task's end time log.
func (r *RPCTask) EndTime(t time.Time) {
	r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// Outputs updates the task's output file log.
func (r *RPCTask) Outputs(f []*tes.OutputFileLog) {
	r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Outputs: f,
		},
	})
}

// Metadata updates the task's metadata log.
func (r *RPCTask) Metadata(m map[string]string) {
	r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
		Id: r.taskID,
		TaskLog: &tes.TaskLog{
			Metadata: m,
		},
	})
}

// ExecutorStartTime updates an executor's start time log.
func (r *RPCTask) ExecutorStartTime(i int, t time.Time) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			StartTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorEndTime updates an executor's end time log.
func (r *RPCTask) ExecutorEndTime(i int, t time.Time) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			EndTime: t.Format(time.RFC3339),
		},
	})
}

// ExecutorExitCode updates an executor's exit code log.
func (r *RPCTask) ExecutorExitCode(i int, x int) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			ExitCode: int32(x),
		},
	})
}

// ExecutorPorts updates an executor's ports log.
func (r *RPCTask) ExecutorPorts(i int, ports []*tes.Ports) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Ports: ports,
		},
	})
}

// ExecutorHostIP updates an executor's host IP log.
func (r *RPCTask) ExecutorHostIP(i int, ip string) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			HostIp: ip,
		},
	})
}

// AppendExecutorStdout appends to an executor's stdout log.
func (r *RPCTask) AppendExecutorStdout(i int, s string) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stdout: s,
		},
	})
}

// AppendExecutorStderr appends to an executor's stderr log.
func (r *RPCTask) AppendExecutorStderr(i int, s string) {
	r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
		Id:   r.taskID,
		Step: int64(i),
		Log: &tes.ExecutorLog{
			Stderr: s,
		},
	})
}

func (r *RPCTask) updateExecutorLogs(up *tl.UpdateExecutorLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateExecutorLogs(ctx, up)
	if err != nil {
		log.Error("Couldn't update executor logs", err)
	}
	cleanup()
	return err
}

func (r *RPCTask) updateTaskLogs(up *tl.UpdateTaskLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateTaskLogs(ctx, up)
	if err != nil {
		log.Error("Couldn't update task logs", err)
	}
	cleanup()
	return err
}

type taskClient struct {
	tes.TaskServiceClient
	tl.TaskLoggerServiceClient
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
	s := tl.NewTaskLoggerServiceClient(conn)
	return &taskClient{t, s}, nil
}
