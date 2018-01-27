package mongodb

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"gopkg.in/mgo.v2/bson"
)

// WriteEvent creates an event for the server to handle.
func (db *MongoDB) WriteEvent(ctx context.Context, req *events.Event) error {
	var err error

	switch req.Type {
	case events.Type_CREATED:
		task := req.GetTask()
		err := db.tasks.Insert(task)
		if err != nil {
			return fmt.Errorf("failed to write task to db: %v", err)
		}

	case events.Type_STATE:
		res, err := db.GetTask(ctx, &tes.GetTaskRequest{
			Id:   req.Id,
			View: tes.TaskView_MINIMAL,
		})
		if err != nil {
			return err
		}
		from := res.State
		to := req.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		err = db.tasks.Update(bson.M{"id": req.Id}, bson.M{"$set": bson.M{"state": to}})

	case events.Type_START_TIME:
		startTime := req.GetStartTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"starttime": startTime}},
		)

	case events.Type_END_TIME:
		endTime := req.GetEndTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"endtime": endTime}},
		)

	case events.Type_OUTPUTS:
		outputs := req.GetOutputs().Value
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"outputs": outputs}},
		)

	case events.Type_METADATA:
		metadata := req.GetMetadata().Value
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"metadata": metadata}},
		)

	case events.Type_EXIT_CODE:
		exitCode := req.GetExitCode()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"exitcode": exitCode}},
		)

	case events.Type_STDOUT:
		stdout := req.GetStdout()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"stdout": stdout}},
		)

	case events.Type_STDERR:
		stderr := req.GetStderr()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{"stderr": stderr}},
		)

	case events.Type_SYSTEM_LOG:
		syslog := req.SysLogString()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$push": bson.M{"systemlogs": syslog}},
		)
	}

	return err
}
