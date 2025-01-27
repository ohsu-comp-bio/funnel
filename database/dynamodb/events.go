package dynamodb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util"
)

// WriteEvent creates an event for the server to handle.
func (db *DynamoDB) WriteEvent(ctx context.Context, e *events.Event) error {
	item := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(e.Id),
			},
		},
	}

	var updateExpr expression.UpdateBuilder

	switch e.Type {
	case events.Type_TASK_CREATED:
		return db.createTask(ctx, e.GetTask())

	case events.Type_TASK_STATE:
		retrier := util.NewRetrier()
		retrier.ShouldRetry = func(err error) bool {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException ||
					aerr.Code() == dynamodb.ErrCodeProvisionedThroughputExceededException {
					return true
				}
			}
			return false
		}

		return retrier.Retry(ctx, func() error {
			// get current state & version
			current := make(map[string]interface{})
			response, err := db.getBasicView(ctx, e.Id)
			if err != nil {
				return err
			}
			if response.Item == nil {
				return tes.ErrNotFound
			}

			err = dynamodbattribute.UnmarshalMap(response.Item, &current)
			if err != nil {
				return fmt.Errorf("failed to DynamoDB unmarshal Task, %v", err)
			}

			// validate state transition
			from := tes.State(current["state"].(float64))
			to := e.GetState()
			if err := tes.ValidateTransition(from, to); err != nil {
				return err
			}

			// apply version restriction and set update
			condExpr := expression.Name("version").Equal(expression.Value(fmt.Sprintf("%v", current["version"])))
			updateExpr = expression.Set(expression.Name("state"), expression.Value(to))
			updateExpr = updateExpr.Set(expression.Name("version"), expression.Value(strconv.FormatInt(time.Now().UnixNano(), 10)))

			// build update item
			expr, err := expression.NewBuilder().WithUpdate(updateExpr).WithCondition(condExpr).Build()
			if err != nil {
				return err
			}

			item.ExpressionAttributeNames = expr.Names()
			item.ExpressionAttributeValues = expr.Values()
			item.UpdateExpression = expr.Update()
			item.ConditionExpression = expr.Condition()

			// apply update with retries upon version collisions
			_, err = db.client.UpdateItemWithContext(ctx, item)
			return err
		})

	case events.Type_TASK_START_TIME:
		if err := db.ensureTaskLog(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].start_time", e.Attempt)),
			expression.Value(e.GetStartTime()),
		)

	case events.Type_TASK_END_TIME:
		if err := db.ensureTaskLog(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].end_time", e.Attempt)),
			expression.Value(e.GetEndTime()),
		)

	case events.Type_TASK_OUTPUTS:
		if err := db.ensureTaskLog(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].outputs", e.Attempt)),
			expression.Value(e.GetOutputs().Value),
		)

	case events.Type_TASK_METADATA:
		if err := db.ensureTaskLog(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		if err := db.ensureTaskLogMetadata(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		for k, v := range e.GetMetadata().Value {
			updateExpr = updateExpr.Set(
				expression.Name(fmt.Sprintf("logs[%v].metadata.%s", e.Attempt, k)),
				expression.Value(v),
			)
		}

	case events.Type_EXECUTOR_START_TIME:
		if err := db.ensureExecLog(ctx, e.Id, e.Attempt, e.Index); err != nil {
			return err
		}
		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].logs[%v].start_time", e.Attempt, e.Index)),
			expression.Value(e.GetStartTime()),
		)

	case events.Type_EXECUTOR_END_TIME:
		if err := db.ensureExecLog(ctx, e.Id, e.Attempt, e.Index); err != nil {
			return err
		}
		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].logs[%v].end_time", e.Attempt, e.Index)),
			expression.Value(e.GetEndTime()),
		)

	case events.Type_EXECUTOR_EXIT_CODE:
		if err := db.ensureExecLog(ctx, e.Id, e.Attempt, e.Index); err != nil {
			return err
		}
		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].logs[%v].exit_code", e.Attempt, e.Index)),
			expression.Value(e.GetExitCode()),
		)

	case events.Type_EXECUTOR_STDOUT:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.stdoutTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
				"attempt_index": {
					S: aws.String(fmt.Sprintf("%v-%v", e.Attempt, e.Index)),
				},
			},
		}

		updateExpr = expression.Set(
			expression.Name("stdout"),
			expression.Value(e.GetStdout()),
		).Set(
			expression.Name("attempt"),
			expression.Value(e.Attempt),
		).Set(
			expression.Name("index"),
			expression.Value(e.Index),
		)

	case events.Type_EXECUTOR_STDERR:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.stderrTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
				"attempt_index": {
					S: aws.String(fmt.Sprintf("%v-%v", e.Attempt, e.Index)),
				},
			},
		}

		updateExpr = expression.Set(
			expression.Name("stderr"),
			expression.Value(e.GetStderr()),
		).Set(
			expression.Name("attempt"),
			expression.Value(e.Attempt),
		).Set(
			expression.Name("index"),
			expression.Value(e.Index),
		)

	case events.Type_SYSTEM_LOG:
		if err := db.ensureSysLog(ctx, e.Id, e.Attempt); err != nil {
			return err
		}

		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.syslogsTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
				"attempt": {
					N: aws.String(strconv.Itoa(int(e.Attempt))),
				},
			},
		}

		updateExpr = expression.Set(
			expression.Name("system_logs"),
			expression.ListAppend(expression.Name("system_logs"), expression.Value([]string{e.SysLogString()})),
		)
	}

	expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()
	if err != nil {
		return err
	}

	item.ExpressionAttributeNames = expr.Names()
	item.ExpressionAttributeValues = expr.Values()
	item.UpdateExpression = expr.Update()
	item.ConditionExpression = expr.Condition()

	_, err = db.client.UpdateItemWithContext(ctx, item)
	return checkErrNotFound(err)
}

func (db *DynamoDB) ensureTaskLog(ctx context.Context, id string, attempt uint32) error {
	// create the log structure for the attempt if it doesnt already exist
	attemptItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET logs[%v] = if_not_exists(logs[%v], :v)", attempt, attempt)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				M: map[string]*dynamodb.AttributeValue{},
			},
		},
	}

	_, err := db.client.UpdateItemWithContext(ctx, attemptItem)
	return checkErrNotFound(err)
}

func (db *DynamoDB) ensureTaskLogMetadata(ctx context.Context, id string, attempt uint32) error {
	// create the log structure for the attempt if it doesnt already exist
	attemptItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET logs[%v].metadata = if_not_exists(logs[%v].metadata, :v)", attempt, attempt)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				M: map[string]*dynamodb.AttributeValue{},
			},
		},
	}

	_, err := db.client.UpdateItemWithContext(ctx, attemptItem)
	return checkErrNotFound(err)
}

func (db *DynamoDB) ensureExecLog(ctx context.Context, id string, attempt, index uint32) error {
	// create the log structure for the executor if it doesnt already exist
	indexItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET logs[%v].logs[%v] = if_not_exists(logs[%v].logs[%v], :v)", attempt, index, attempt, index)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				M: map[string]*dynamodb.AttributeValue{},
			},
		},
	}
	_, err := db.client.UpdateItemWithContext(ctx, indexItem)
	return checkErrNotFound(err)
}

func (db *DynamoDB) ensureSysLog(ctx context.Context, id string, attempt uint32) error {
	item := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.syslogsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
			"attempt": {
				N: aws.String(strconv.Itoa(int(attempt))),
			},
		},
		UpdateExpression: aws.String("SET system_logs = if_not_exists(system_logs, :v)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				L: []*dynamodb.AttributeValue{},
			},
		},
	}

	_, err := db.client.UpdateItemWithContext(ctx, item)
	return checkErrNotFound(err)
}

func checkErrNotFound(err error) error {
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			return tes.ErrNotFound
		}
	}
	return err
}
