package badger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/util"
)

// WriteEvent creates an event for the server to handle.
func (db *Badger) WriteEvent(ctx context.Context, req *events.Event) error {
	retrier := util.NewRetrier()
	retrier.ShouldRetry = func(err error) bool {
		return true
	}

	return retrier.Retry(ctx, func() error {
		return db.writeEvent(ctx, req)
	})
}

func (db *Badger) writeEvent(ctx context.Context, req *events.Event) error {
	err := db.db.Update(func(txn *badger.Txn) error {

		// If this event creates a new task, we don't need to update logic below,
		// just marshal and save the task.
		if req.Type == events.Type_TASK_CREATED {
			task := req.GetTask()
			val, err := proto.Marshal(task)
			if err != nil {
				return fmt.Errorf("marshaling task to bytes: %s", err)
			}

			return txn.Set(taskKey(task.Id), val)
		}

		// The rest of the events below all update a task, so we need to make sure it exists.
		task, err := db.getTask(txn, req.Id)
		if err != nil {
			return err
		}

		switch req.Type {
		case events.Type_TASK_STATE:
			task.State = req.GetState()

		case events.Type_TASK_START_TIME:
			task.GetTaskLog(0).StartTime = req.GetStartTime()

		case events.Type_TASK_END_TIME:
			task.GetTaskLog(0).EndTime = req.GetEndTime()

		case events.Type_TASK_OUTPUTS:
			task.GetTaskLog(0).Outputs = req.GetOutputs().Value

		case events.Type_TASK_METADATA:
			task.GetTaskLog(0).Metadata = req.GetMetadata().Value

		case events.Type_EXECUTOR_START_TIME:
			task.GetExecLog(0, int(req.Index)).StartTime = req.GetStartTime()

		case events.Type_EXECUTOR_END_TIME:
			task.GetExecLog(0, int(req.Index)).EndTime = req.GetEndTime()

		case events.Type_EXECUTOR_EXIT_CODE:
			task.GetExecLog(0, int(req.Index)).ExitCode = req.GetExitCode()

		case events.Type_EXECUTOR_STDOUT:
			task.GetExecLog(0, int(req.Index)).Stdout = req.GetStdout()

		case events.Type_EXECUTOR_STDERR:
			task.GetExecLog(0, int(req.Index)).Stderr = req.GetStderr()

		case events.Type_SYSTEM_LOG:
			tl := task.GetTaskLog(0)
			tl.SystemLogs = append(tl.SystemLogs, req.SysLogString())
		}

		val, err := proto.Marshal(task)
		if err != nil {
			return fmt.Errorf("marshaling task to bytes: %s", err)
		}

		return txn.Set(taskKey(task.Id), val)
	})

	if err != nil {
		return fmt.Errorf("storing task in database: %s", err)
	}
	return nil
}
