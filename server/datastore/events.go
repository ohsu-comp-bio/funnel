package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func (d *Datastore) WriteEvent(ctx context.Context, e *events.Event) error {

	switch e.Type {

	case events.Type_TASK_CREATED:
		putKeys, putData := marshalTask(e.GetTask())
		_, err := d.client.PutMulti(ctx, putKeys, putData)
		if err != nil {
			return err
		}

	case events.Type_EXECUTOR_STDOUT:
		_, err := d.client.Put(ctx, stdoutKey(e.Id, e.Attempt, e.Index), marshalEvent(e))
		return err

	case events.Type_EXECUTOR_STDERR:
		_, err := d.client.Put(ctx, stderrKey(e.Id, e.Attempt, e.Index), marshalEvent(e))
		return err

	case events.Type_TASK_STATE:
		res, err := d.GetTask(ctx, &tes.GetTaskRequest{
			Id: e.Id,
		})
		if err != nil {
			return err
		}

		from := res.State
		to := e.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		fallthrough

	default:
		_, err := d.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			props := datastore.PropertyList{}
			err := tx.Get(taskKey(e.Id), &props)
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

			putKeys, putData := marshalTask(task)
			_, err = tx.PutMulti(putKeys, putData)
			return err
		})
		return err
	}
	return nil
}
