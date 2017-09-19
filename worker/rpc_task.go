package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
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

func (r *RPCTask) Write(ev *events.Event) error {
	switch ev.Type {
	case events.Type_STATE:
		_, err := r.client.UpdateTaskState(context.Background(), &tl.UpdateTaskStateRequest{
			Id:    r.taskID,
			State: ev.State,
		})
		return err
	case events.Type_START_TIME:
		return r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
			Id: r.taskID,
			TaskLog: &tes.TaskLog{
				StartTime: events.TimestampString(ev.StartTime),
			},
		})
	case events.Type_END_TIME:
		return r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
			Id: r.taskID,
			TaskLog: &tes.TaskLog{
				EndTime: events.TimestampString(ev.EndTime),
			},
		})
	case events.Type_OUTPUTS:
		return r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
			Id: r.taskID,
			TaskLog: &tes.TaskLog{
				Outputs: ev.Outputs,
			},
		})
	case events.Type_METADATA:
		return r.updateTaskLogs(&tl.UpdateTaskLogsRequest{
			Id: r.taskID,
			TaskLog: &tes.TaskLog{
				Metadata: ev.Metadata,
			},
		})
	case events.Type_EXECUTOR_START_TIME:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				StartTime: events.TimestampString(ev.ExecutorStartTime),
			},
		})
	case events.Type_EXECUTOR_END_TIME:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				EndTime: events.TimestampString(ev.ExecutorEndTime),
			},
		})
	case events.Type_EXIT_CODE:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				ExitCode: ev.ExitCode,
			},
		})
	case events.Type_PORTS:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				Ports: ev.Ports,
			},
		})
	case events.Type_HOST_IP:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				HostIp: ev.HostIp,
			},
		})
	case events.Type_STDOUT:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				Stdout: ev.Stdout,
			},
		})
	case events.Type_STDERR:
		return r.updateExecutorLogs(&tl.UpdateExecutorLogsRequest{
			Id:   r.taskID,
			Step: int64(ev.Index),
			Log: &tes.ExecutorLog{
				Stderr: ev.Stderr,
			},
		})
	}
	return nil
}

func (r *RPCTask) updateExecutorLogs(up *tl.UpdateExecutorLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateExecutorLogs(ctx, up)
	cleanup()
	return err
}

func (r *RPCTask) updateTaskLogs(up *tl.UpdateTaskLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), r.updateTimeout)
	_, err := r.client.UpdateTaskLogs(ctx, up)
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
