package elastic

import (
	"bytes"
	"context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	elastic "gopkg.in/olivere/elastic.v5"
)

var updateTaskLogs = `
if (ctx._source.logs == null) {
  ctx._source.logs = new ArrayList();
}

// Ensure the task logs array is long enough.
for (; params.attempt > ctx._source.logs.length - 1; ) {
  Map m = new HashMap();
  m.logs = new ArrayList();
  ctx._source.logs.add(m);
}

// Set the field.
if (params.field == "system_logs") {
  if (ctx._source.logs[params.attempt].system_logs == null) {
    ctx._source.logs[params.attempt].system_logs = new ArrayList();
  }
  ctx._source.logs[params.attempt].system_logs.add(params.value)
} else if (params.field == "metadata") {
  if (ctx._source.logs[params.attempt].metadata == null) {
    ctx._source.logs[params.attempt].metadata = new HashMap();
  }
  for (entry in params.value.entrySet()) {
    ctx._source.logs[params.attempt].metadata.put(entry.getKey(), entry.getValue())
  }
} else {
  ctx._source.logs[params.attempt][params.field] = params.value;
}
`

var updateExecutorLogs = `
if (ctx._source.logs == null) {
  ctx._source.logs = new ArrayList();
}

// Ensure the task logs array is long enough.
for (; params.attempt > ctx._source.logs.length - 1; ) {
  Map m = new HashMap();
  m.logs = new ArrayList();
  ctx._source.logs.add(m);
}

// Ensure the executor logs array is long enough.
for (; params.index > ctx._source.logs[params.attempt].logs.length - 1; ) {
  Map m = new HashMap();
  ctx._source.logs[params.attempt].logs.add(m);
}

// Set the field.
ctx._source.logs[params.attempt].logs[params.index][params.field] = params.value;
`

func taskLogUpdate(attempt uint32, field string, value interface{}) *elastic.Script {
	return elastic.NewScript(updateTaskLogs).
		Lang("painless").
		Param("attempt", attempt).
		Param("field", field).
		Param("value", value)
}

func execLogUpdate(attempt, index uint32, field string, value interface{}) *elastic.Script {
	return elastic.NewScript(updateExecutorLogs).
		Lang("painless").
		Param("attempt", attempt).
		Param("index", index).
		Param("field", field).
		Param("value", value)
}

// WriteEvent writes a task update event.
func (es *Elastic) WriteEvent(ctx context.Context, ev *events.Event) error {
	u := es.client.Update().
		Index(es.taskIndex).
		Type("task").
		RetryOnConflict(10).
		Id(ev.Id)

	switch ev.Type {
	case events.Type_TASK_CREATED:
		task := ev.GetTask()
		mar := jsonpb.Marshaler{}
		s, err := mar.MarshalToString(task)
		if err != nil {
			return err
		}

		_, err = es.client.Index().
			Index(es.taskIndex).
			Type("task").
			Id(task.Id).
			BodyString(s).
			Do(ctx)
		return err

	case events.Type_TASK_STATE:
		retrier := util.NewRetrier()
		retrier.ShouldRetry = func(err error) bool {
			if elastic.IsConflict(err) || elastic.IsConnErr(err) {
				return true
			}
			return false
		}

		return retrier.Retry(ctx, func() error {
			// get current state & version
			res, err := es.getTask(ctx, &tes.GetTaskRequest{Id: ev.Id})
			if err != nil {
				return err
			}

			task := &tes.Task{}
			err = jsonpb.Unmarshal(bytes.NewReader(*res.Source), task)
			if err != nil {
				return err
			}

			// validate state transition
			from := task.State
			to := ev.GetState()
			if err := tes.ValidateTransition(from, to); err != nil {
				return err
			}

			// apply version restriction and set update
			_, err = es.client.Update().
				Index(es.taskIndex).
				Type("task").
				Id(ev.Id).
				Version(*res.Version).
				Doc(map[string]string{"state": to.String()}).
				Do(ctx)
			return err
		})

	case events.Type_TASK_START_TIME:
		u = u.Script(taskLogUpdate(ev.Attempt, "start_time", ev.GetStartTime()))

	case events.Type_TASK_END_TIME:
		u = u.Script(taskLogUpdate(ev.Attempt, "end_time", ev.GetEndTime()))

	case events.Type_TASK_OUTPUTS:
		u = u.Script(taskLogUpdate(ev.Attempt, "outputs", ev.GetOutputs().Value))

	case events.Type_TASK_METADATA:
		u = u.Script(taskLogUpdate(ev.Attempt, "metadata", ev.GetMetadata().Value))

	case events.Type_EXECUTOR_START_TIME:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "start_time", ev.GetStartTime()))

	case events.Type_EXECUTOR_END_TIME:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "end_time", ev.GetEndTime()))

	case events.Type_EXECUTOR_EXIT_CODE:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "exit_code", ev.GetExitCode()))

	case events.Type_EXECUTOR_STDOUT:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "stdout", ev.GetStdout()))

	case events.Type_EXECUTOR_STDERR:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "stderr", ev.GetStderr()))

	case events.Type_SYSTEM_LOG:
		u = u.Script(taskLogUpdate(ev.Attempt, "system_logs", ev.SysLogString()))
	}

	_, err := u.Do(ctx)
	if elastic.IsNotFound(err) {
		return tes.ErrNotFound
	}
	return err
}
