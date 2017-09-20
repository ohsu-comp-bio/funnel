package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

// CreateEvent creates an event for the server to handle.
func (taskBolt *TaskBolt) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	var err error

	tl := &tes.TaskLog{}
	el := &tes.ExecutorLog{}

	// used to trim stdout and stderr in ExecutorLog
	max := taskBolt.conf.Server.MaxExecutorLogSize

	switch req.Type {
	case events.Type_TASK_STATE:
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return transitionTaskState(tx, req.Id, req.GetState())
		})

	case events.Type_TASK_START_TIME:
		tl.StartTime = ptypes.TimestampString(req.GetStartTime())
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateTaskLogs(tx, req.Id, tl)
		})

	case events.Type_TASK_END_TIME:
		tl.EndTime = ptypes.TimestampString(req.GetEndTime())
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateTaskLogs(tx, req.Id, tl)
		})

	case events.Type_TASK_OUTPUTS:
		tl.Outputs = req.GetOutputs().Value
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateTaskLogs(tx, req.Id, tl)
		})

	case events.Type_TASK_METADATA:
		tl.Metadata = req.GetMetadata().Value
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateTaskLogs(tx, req.Id, tl)
		})

	case events.Type_EXECUTOR_START_TIME:
		el.StartTime = ptypes.TimestampString(req.GetStartTime())
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_END_TIME:
		el.EndTime = ptypes.TimestampString(req.GetEndTime())
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_EXIT_CODE:
		el.ExitCode = req.GetExitCode()
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_HOST_IP:
		el.HostIp = req.GetHostIp()
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_PORTS:
		el.Ports = req.GetPorts().Value
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_STDOUT:
		el.Stdout = req.GetStdout()
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})

	case events.Type_EXECUTOR_STDERR:
		el.Stderr = req.GetStderr()
		err = taskBolt.db.Update(func(tx *bolt.Tx) error {
			return updateExecutorLogs(tx, fmt.Sprint(req.Id, req.Index), el, max)
		})
	}

	if err != nil {
		return nil, err
	}

	return &events.CreateEventResponse{}, nil
}

func updateTaskLogs(tx *bolt.Tx, id string, tl *tes.TaskLog) error {
	tasklog := &tes.TaskLog{}

	// Try to load existing task log
	b := tx.Bucket(TasksLog).Get([]byte(id))
	if b != nil {
		err := proto.Unmarshal(b, tasklog)
		if err != nil {
			return err
		}
	}

	if tl.StartTime != "" {
		tasklog.StartTime = tl.StartTime
	}

	if tl.EndTime != "" {
		tasklog.EndTime = tl.EndTime
	}

	if tl.Outputs != nil {
		tasklog.Outputs = tl.Outputs
	}

	if tl.Metadata != nil {
		if tasklog.Metadata == nil {
			tasklog.Metadata = map[string]string{}
		}
		for k, v := range tl.Metadata {
			tasklog.Metadata[k] = v
		}
	}

	logbytes, err := proto.Marshal(tasklog)
	if err != nil {
		return err
	}
	return tx.Bucket(TasksLog).Put([]byte(id), logbytes)
}

func updateExecutorLogs(tx *bolt.Tx, id string, el *tes.ExecutorLog, max int) error {
	// Check if there is an existing task log
	o := tx.Bucket(ExecutorLogs).Get([]byte(id))
	if o != nil {
		// There is an existing log in the DB, load it
		existing := &tes.ExecutorLog{}
		// max bytes to be stored in the db
		err := proto.Unmarshal(o, existing)
		if err != nil {
			return err
		}

		stdout := []byte(existing.Stdout + el.Stdout)
		stderr := []byte(existing.Stderr + el.Stderr)

		// Trim the stdout/err logs to the max size if needed
		if len(stdout) > max {
			stdout = stdout[:max]
		}
		if len(stderr) > max {
			stderr = stderr[:max]
		}

		el.Stdout = string(stdout)
		el.Stderr = string(stderr)

		// Merge the updates into the existing.
		proto.Merge(existing, el)
		// existing is updated, so set that to el which will get saved below.
		el = existing
	}

	// Save the updated log
	logbytes, err := proto.Marshal(el)
	if err != nil {
		return err
	}
	return tx.Bucket(ExecutorLogs).Put([]byte(id), logbytes)
}
