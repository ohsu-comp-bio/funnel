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
	case events.Type_TASK_CREATED:
		task := req.GetTask()
		task.Logs = []*tes.TaskLog{
			{
				Logs: []*tes.ExecutorLog{},
			},
		}

		err := db.tasks.Insert(task)
		if err != nil {
			return fmt.Errorf("failed to write task to db: %v", err)
		}

	case events.Type_TASK_STATE:
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

	case events.Type_TASK_START_TIME:
		startTime := req.GetStartTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.starttime", req.Attempt): startTime}},
		)

	case events.Type_TASK_END_TIME:
		endTime := req.GetEndTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.endtime", req.Attempt): endTime}},
		)

	case events.Type_TASK_OUTPUTS:
		outputs := req.GetOutputs().Value
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.outputs", req.Attempt): outputs}},
		)

	case events.Type_TASK_METADATA:
		metadata := req.GetMetadata().Value
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.metadata", req.Attempt): metadata}},
		)

	case events.Type_EXECUTOR_START_TIME:
		startTime := req.GetStartTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.starttime", req.Attempt, req.Index): startTime}},
		)

	case events.Type_EXECUTOR_END_TIME:
		endTime := req.GetEndTime()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.endtime", req.Attempt, req.Index): endTime}},
		)

	case events.Type_EXECUTOR_EXIT_CODE:
		exitCode := req.GetExitCode()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.exitcode", req.Attempt, req.Index): exitCode}},
		)

	case events.Type_EXECUTOR_STDOUT:
		stdout := req.GetStdout()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.stdout", req.Attempt, req.Index): stdout}},
		)

	case events.Type_EXECUTOR_STDERR:
		stderr := req.GetStderr()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.stderr", req.Attempt, req.Index): stderr}},
		)

	case events.Type_SYSTEM_LOG:
		syslog := req.GetSystemLog().LogString()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$push": bson.M{fmt.Sprintf("logs.%v.systemlogs", req.Attempt): syslog}},
		)
	}

	return err
}
