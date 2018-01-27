package elastic

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	elastic "gopkg.in/olivere/elastic.v5"
)

var updateTaskLogs = `

// Set the field.
if (params.field == "system_logs") {
  if (ctx._source.system_logs == null) {
    ctx._source.system_logs = new ArrayList();
  }
  ctx._source.system_logs.add(params.value)
} else {
  ctx._source[params.field] = params.value;
}
`

func taskLogUpdate(attempt uint32, field string, value interface{}) *elastic.Script {
	return elastic.NewScript(updateTaskLogs).
		Lang("painless").
		Param("attempt", attempt).
		Param("field", field).
		Param("value", value)
}

// WriteEvent writes a task update event.
func (es *Elastic) WriteEvent(ctx context.Context, ev *events.Event) error {
	u := es.client.Update().
		Index(es.taskIndex).
		Type("task").
		RetryOnConflict(3).
		Id(ev.Id)

	switch ev.Type {
	case events.Type_CREATED:
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

	case events.Type_STATE:
		res, err := es.GetTask(ctx, &tes.GetTaskRequest{
			Id: ev.Id,
		})
		if err != nil {
			return err
		}

		from := res.State
		to := ev.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		u = u.Doc(map[string]string{"state": to.String()})

	case events.Type_START_TIME:
		u = u.Script(taskLogUpdate("start_time", ev.GetStartTime()))

	case events.Type_END_TIME:
		u = u.Script(taskLogUpdate("end_time", ev.GetEndTime()))

	case events.Type_OUTPUTS:
		u = u.Script(taskLogUpdate("outputs", ev.GetOutputs().Value))

	case events.Type_METADATA:
		u = u.Script(taskLogUpdate("metadata", ev.GetMetadata().Value))

	case events.Type_EXIT_CODE:
		u = u.Script(execLogUpdate("exit_code", ev.GetExitCode()))

	case events.Type_STDOUT:
		u = u.Script(execLogUpdate("stdout", ev.GetStdout()))

	case events.Type_STDERR:
		u = u.Script(execLogUpdate("stderr", ev.GetStderr()))

	case events.Type_SYSTEM_LOG:
		u = u.Script(taskLogUpdate("system_logs", ev.SysLogString()))
	}

	_, err := u.Do(ctx)
	if elastic.IsNotFound(err) {
		return tes.ErrNotFound
	}
	return err
}
