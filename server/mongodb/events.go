package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// WriteEvent creates an event for the server to handle.
func (db *MongoDB) WriteEvent(ctx context.Context, req *events.Event) error {
	update := bson.M{}
	selector := bson.M{"id": req.Id}

	switch req.Type {
	case events.Type_TASK_CREATED:
		task := req.GetTask()
		task.Logs = []*tes.TaskLog{
			{
				Logs: []*tes.ExecutorLog{},
			},
		}
		return db.tasks.Insert(&task)

	case events.Type_TASK_STATE:
		retrier := util.NewRetrier()
		retrier.ShouldRetry = func(err error) bool {
			if err == mgo.ErrNotFound {
				return true
			}
			return false
		}

		return retrier.Retry(ctx, func() error {
			// get current state & version
			current := make(map[string]interface{})
			q := db.tasks.Find(bson.M{"id": req.Id}).Select(bson.M{"state": 1, "version": 1})
			err := q.One(&current)
			if err == mgo.ErrNotFound {
				return tes.ErrNotFound
			}
			if err != nil {
				return err
			}

			// validate state transition
			from := tes.State(current["state"].(int))
			to := req.GetState()
			if err = tes.ValidateTransition(from, to); err != nil {
				return err
			}

			// apply version restriction and set update
			selector["version"] = current["version"]
			update = bson.M{"$set": bson.M{"state": to, "version": time.Now().UnixNano()}}
			return db.tasks.Update(selector, update)
		})

	case events.Type_TASK_START_TIME:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.starttime", req.Attempt): req.GetStartTime(),
			},
		}

	case events.Type_TASK_END_TIME:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.endtime", req.Attempt): req.GetEndTime(),
			},
		}

	case events.Type_TASK_OUTPUTS:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.outputs", req.Attempt): req.GetOutputs().Value,
			},
		}

	case events.Type_TASK_METADATA:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.metadata", req.Attempt): req.GetMetadata().Value,
			},
		}

	case events.Type_EXECUTOR_START_TIME:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.logs.%v.starttime", req.Attempt, req.Index): req.GetStartTime(),
			},
		}

	case events.Type_EXECUTOR_END_TIME:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.logs.%v.endtime", req.Attempt, req.Index): req.GetEndTime(),
			},
		}

	case events.Type_EXECUTOR_EXIT_CODE:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.logs.%v.exitcode", req.Attempt, req.Index): req.GetExitCode(),
			},
		}

	case events.Type_EXECUTOR_STDOUT:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.logs.%v.stdout", req.Attempt, req.Index): req.GetStdout(),
			},
		}

	case events.Type_EXECUTOR_STDERR:
		update = bson.M{
			"$set": bson.M{
				fmt.Sprintf("logs.%v.logs.%v.stderr", req.Attempt, req.Index): req.GetStderr(),
			},
		}

	case events.Type_SYSTEM_LOG:
		update = bson.M{
			"$push": bson.M{
				fmt.Sprintf("logs.%v.systemlogs", req.Attempt): req.SysLogString(),
			},
		}
	}

	return db.tasks.Update(selector, update)
}
