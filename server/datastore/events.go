package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func (d *Datastore) WriteEvent(ctx context.Context, e *events.Event) error {
	taskKey := datastore.NameKey("Task", e.Id, nil)
	// TODO
	//contentKey := datastore.NameKey("TaskChunk", "0-content", taskKey)

	switch e.Type {

	case events.Type_TASK_CREATED:
		_, err := d.client.Put(ctx, taskKey, marshalTask(e.GetTask()))
		if err != nil {
			return err
		}

	case events.Type_EXECUTOR_STDOUT:
		_, err := d.client.Put(ctx, stdoutKey(taskKey, e.Attempt, e.Index), marshalEvent(e))
		return err

	case events.Type_EXECUTOR_STDERR:
		_, err := d.client.Put(ctx, stderrKey(taskKey, e.Attempt, e.Index), marshalEvent(e))
		return err

	default:
		_, err := d.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			props := datastore.PropertyList{}
			err := tx.Get(taskKey, &props)
			if err != nil {
				return err
			}

			task := &tes.Task{}
			unmarshalTask(task, props)
			tb := events.TaskBuilder{task}
			err = tb.WriteEvent(context.Background(), e)
			if err != nil {
				return err
			}

			_, err = tx.Put(taskKey, marshalTask(task))
			return err
		})
		return err
	}
	return nil
}
