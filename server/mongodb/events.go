package mongodb

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

// CreateEvent creates an event for the server to handle.
func (db *MongoDB) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	return nil, fmt.Errorf("CreateEvent - Not Implemented")
}

// Write writes task events to the database, updating the task record they
// are related to. System log events are ignored.
func (db *MongoDB) Write(req *events.Event) error {
	return db.WriteContext(context.Background(), req)
}

// WriteContext is Write, but with context.
func (db *MongoDB) WriteContext(ctx context.Context, req *events.Event) error {
	var err error

	switch req.Type {
	case events.Type_TASK_STATE:
		res, err := db.GetTask(ctx, &tes.GetTaskRequest{
			Id:   req.Id,
			View: tes.TaskView_MINIMAL,
		})
		if err != nil {
			return fmt.Errorf("error fetch current state: %v", err)
		}
		from := res.State
		to := req.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		err = db.tasks.Update(bson.M{"id": req.Id}, bson.M{"$set": bson.M{"state": to}})

	case events.Type_TASK_START_TIME:
		startTime := ptypes.TimestampString(req.GetStartTime())
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.starttime", req.Attempt): startTime}},
		)

	case events.Type_TASK_END_TIME:
		endTime := ptypes.TimestampString(req.GetEndTime())
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
		startTime := ptypes.TimestampString(req.GetStartTime())
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.starttime", req.Attempt, req.Index): startTime}},
		)

	case events.Type_EXECUTOR_END_TIME:
		endTime := ptypes.TimestampString(req.GetEndTime())
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

	case events.Type_EXECUTOR_HOST_IP:
		hostIP := req.GetHostIp()
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.hostip", req.Attempt, req.Index): hostIP}},
		)

	case events.Type_EXECUTOR_PORTS:
		ports := req.GetPorts().Value
		err = db.tasks.Update(
			bson.M{"id": req.Id},
			bson.M{"$set": bson.M{fmt.Sprintf("logs.%v.logs.%v.ports", req.Attempt, req.Index): ports}},
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
	}

	return err
}
