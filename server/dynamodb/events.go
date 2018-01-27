package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strconv"
)

// WriteEvent creates an event for the server to handle.
func (db *DynamoDB) WriteEvent(ctx context.Context, e *events.Event) error {
	if e.Type == events.Type_CREATED {
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

	switch e.Type {

	case events.Type_STATE:
		task, err := db.GetTask(ctx, &tes.GetTaskRequest{
			Id:   e.Id,
			View: tes.TaskView_MINIMAL,
		})
		if err != nil {
			return err
		}

		from := task.State
		to := e.GetState()
		if err := tes.ValidateTransition(from, to); err != nil {
			return err
		}
		item.ExpressionAttributeNames = map[string]*string{
			"#state": aws.String("state"),
		}
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":to": {
				N: aws.String(strconv.Itoa(int(to))),
			},
		}
		item.UpdateExpression = aws.String("SET #state = :to")

	case events.Type_OUTPUTS:
		val, err := dynamodbattribute.MarshalList(e.GetOutputs().Value)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB marshal TaskLog Outputs, %v", err)
		}
		item.UpdateExpression = aws.String("SET outputs = :c")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				L: val,
			},
		}

	case events.Type_METADATA:
		val, err := dynamodbattribute.MarshalMap(e.GetMetadata().Value)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB marshal TaskLog Metadata, %v", err)
		}

		item.UpdateExpression = aws.String("SET metadata = :c")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				M: val,
			},
		}

	case events.Type_START_TIME:
		item.UpdateExpression = aws.String("SET start_time = :c")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(e.GetStartTime()),
			},
		}

	case events.Type_END_TIME:
		item.UpdateExpression = aws.String("SET end_time = :c")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(e.GetEndTime()),
			},
		}

	case events.Type_EXIT_CODE:
		item.UpdateExpression = aws.String("SET exit_code = :c")
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				N: aws.String(strconv.Itoa(int(e.GetExitCode()))),
			},
		}

	case events.Type_STDOUT:
		stdout := e.GetStdout()
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.stdoutTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
			},
			UpdateExpression: aws.String("SET stdout = :stdout"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":stdout": {
					S: aws.String(stdout),
				},
			},
		}

	case events.Type_STDERR:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.stderrTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
			},
			UpdateExpression: aws.String("SET stderr = :stderr"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":stderr": {
					S: aws.String(e.GetStderr()),
				},
			},
		}

	case events.Type_SYSTEM_LOG:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(db.syslogsTable),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(e.Id),
				},
			},
			UpdateExpression: aws.String("SET system_logs = list_append(if_not_exists(system_logs, :e), :syslog)"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":e": {
					L: []*dynamodb.AttributeValue{},
				},
				":syslog": {
					L: []*dynamodb.AttributeValue{
						{
							S: aws.String(e.SysLogString()),
						},
					},
				},
			},
		}
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
