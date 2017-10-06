package events

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/server/elastic"
	"golang.org/x/net/context"
)

type ElasticWriter struct {
  es *elastic.Elastic
}

// Write writes a task update event.
func (ew *ElasticWriter) Write(ev *Event) error {
	ctx := context.Background()

	switch ev.Type {
	case Type_TASK_CREATED:
		return ew.es.CreateTask(ctx, ev.GetTask())

	case Type_SYSTEM_LOG:
		return ew.es.CreateSyslog(ctx, ev)
	}

	task, err := ew.es.GetTask(ctx, ev.Id)
	if err != nil {
		return err
	}

	err = TaskBuilder{Task: task}.Write(ev)
	if err != nil {
		return err
	}

  return ew.es.UpdateTask(ctx, task)
}
