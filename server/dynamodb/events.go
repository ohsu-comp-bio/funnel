package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// WriteEvent creates an event for the server to handle.
func (db *DynamoDB) WriteEvent(ctx context.Context, e *events.Event) error {
	if e.Type == events.Type_TASK_CREATED {
		return db.createTask(ctx, e.GetTask())
	}

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
	exprBuilder := expression.NewBuilder()

	switch e.Type {

	case events.Type_TASK_STATE:
		response, err := db.getBasicView(ctx, e.Id)
		if err != nil {
			return err
		}

		type altMinView struct {
			Id      string
			State   tes.State
			Version int32
		}

		task := altMinView{}
		err = dynamodbattribute.UnmarshalMap(response.Item, &task)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB unmarshal Task, %v", err)
		}

		from := task.State
		to := e.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}

		exprBuilder = exprBuilder.WithCondition(expression.Name("version").Equal(expression.Value(task.Version)))
		updateExpr = expression.Set(expression.Name("state"), expression.Value(to))
		updateExpr = updateExpr.Add(expression.Name("version"), expression.Value(1))

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

		updateExpr = expression.Set(
			expression.Name(fmt.Sprintf("logs[%v].metadata", e.Attempt)),
			expression.Value(e.GetMetadata().Value),
		)

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

	// ensure item exists
	// exprBuilder = exprBuilder.WithCondition(expression.Name("id").Equal(expression.Value(e.Id)))

	expr, err := exprBuilder.WithUpdate(updateExpr).Build()
	if err != nil {
		return err
	}

	item.ExpressionAttributeNames = expr.Names()
	item.ExpressionAttributeValues = expr.Values()
	item.UpdateExpression = expr.Update()
	if *expr.Condition() != "" {
		item.ConditionExpression = expr.Condition()
	}

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
