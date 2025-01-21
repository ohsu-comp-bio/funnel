package datastore

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// WriteEvent writes a task event to the database.
func (d *Datastore) WriteEvent(ctx context.Context, e *events.Event) error {

	switch e.Type {

	case events.Type_TASK_CREATED:
		putKeys, putData := marshalTask(e.GetTask(), ctx)
		_, err := d.client.PutMulti(ctx, putKeys, putData)
		return err

	case events.Type_EXECUTOR_STDOUT:
		_, err := d.client.Put(ctx, stdoutKey(e.Id, e.Attempt, e.Index), marshalEvent(e))
		return err

	case events.Type_EXECUTOR_STDERR:
		_, err := d.client.Put(ctx, stderrKey(e.Id, e.Attempt, e.Index), marshalEvent(e))
		return err

	case events.Type_SYSTEM_LOG:
		_, err := d.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			props := datastore.PropertyList{}
			err := tx.Get(sysLogsKey(e.Id, e.Attempt), &props)
			if err != nil && err != datastore.ErrNoSuchEntity {
				return err
			}

			p := &part{}
			err = datastore.LoadStruct(p, props)
			if err != nil {
				return err
			}

			_, err = tx.Put(sysLogsKey(e.Id, e.Attempt), &part{
				Type:       sysLogsPart,
				Attempt:    int(e.Attempt),
				Index:      int(e.Index),
				SystemLogs: append(p.SystemLogs, e.SysLogString()),
			})
			return err
		})
		return err

	default:
		_, err := d.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			props := datastore.PropertyList{}
			err := tx.Get(taskKey(e.Id), &props)
			if err == datastore.ErrNoSuchEntity {
				return tes.ErrNotFound
			}
			if err != nil {
				return err
			}

			task := &tes.Task{}
			if err := unmarshalTask(task, props, ctx); err != nil {
				return err
			}

			if e.Type == events.Type_TASK_STATE {
				from := task.State
				to := e.GetState()
				if err := tes.ValidateTransition(from, to); err != nil {
					return err
				}
			}

			tb := events.TaskBuilder{Task: task}
			err = tb.WriteEvent(context.Background(), e)
			if err != nil {
				return err
			}

			putKeys, putData := marshalTask(task, ctx)
			_, err = tx.PutMulti(putKeys, putData)
			return err
		})
		return err
	}
}
