package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"strconv"
)

// CreateEvent creates an event for the server to handle.
func (db *DynamoDB) CreateEvent(ctx context.Context, req *events.Event) (*events.CreateEventResponse, error) {
	err := db.WriteContext(ctx, req)
	return &events.CreateEventResponse{}, err
}

// Write writes task events to the database, updating the task record they
// are related to. System log events are ignored.
func (db *DynamoDB) Write(req *events.Event) error {
	return db.WriteContext(context.Background(), req)
}

// WriteContext is Write, but with context.
func (db *DynamoDB) WriteContext(ctx context.Context, e *events.Event) error {
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
		ReturnValues: aws.String("UPDATED_NDB"),
	}

	// create the log structure for the attempt if it doesnt already exist
	attemptItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(e.Id),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET logs[%v] = if_not_exists(logs[%v], :v)", e.Attempt, e.Attempt)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				M: map[string]*dynamodb.AttributeValue{},
			},
		},
		ReturnValues: aws.String("UPDATED_NDB"),
	}
	_, err := db.client.UpdateItem(attemptItem)
	if err != nil {
		return err
	}

	// create the log structure for the executor if it doesnt already exist
	indexItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(e.Id),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET logs[%v].logs[%v] = if_not_exists(logs[%v].logs[%v], :v)", e.Attempt, e.Index, e.Attempt, e.Index)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				M: map[string]*dynamodb.AttributeValue{},
			},
		},
		ReturnValues: aws.String("UPDATED_NDB"),
	}
	_, err = db.client.UpdateItem(indexItem)
	if err != nil {
		return err
	}

	switch e.Type {
	case events.Type_TASK_STATE:
		item.ExpressionAttributeNames = map[string]*string{
			"#state": aws.String("state"),
		}
		item.UpdateExpression = aws.String("SET #state = :to")

		// define valid transitions
		switch e.GetState() {

		case tes.State_INITIALIZING:
			item.ConditionExpression = aws.String("#state = :from")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(tes.State_INITIALIZING))),
				},
				":from": {
					N: aws.String(strconv.Itoa(int(tes.State_QUEUED))),
				},
			}

		case tes.State_RUNNING:
			item.ConditionExpression = aws.String("#state = :from")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(tes.State_RUNNING))),
				},
				":from": {
					N: aws.String(strconv.Itoa(int(tes.State_INITIALIZING))),
				},
			}

		case tes.State_COMPLETE:
			item.ConditionExpression = aws.String("#state = :from")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(tes.State_COMPLETE))),
				},
				":from": {
					N: aws.String(strconv.Itoa(int(tes.State_RUNNING))),
				},
			}

		case tes.State_EXECUTOR_ERROR:
			item.ConditionExpression = aws.String("#state IN (:i, :r)")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(tes.State_EXECUTOR_ERROR))),
				},
				":i": {
					N: aws.String(strconv.Itoa(int(tes.State_INITIALIZING))),
				},
				":r": {
					N: aws.String(strconv.Itoa(int(tes.State_RUNNING))),
				},
			}

		case tes.State_SYSTEM_ERROR, tes.State_CANCELED:
			item.ConditionExpression = aws.String("#state IN (:q, :i, :r)")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(e.GetState()))),
				},
				":q": {
					N: aws.String(strconv.Itoa(int(tes.State_QUEUED))),
				},
				":i": {
					N: aws.String(strconv.Itoa(int(tes.State_INITIALIZING))),
				},
				":r": {
					N: aws.String(strconv.Itoa(int(tes.State_RUNNING))),
				},
			}
		}

	case events.Type_TASK_START_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].start_time = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetStartTime())),
			},
		}

	case events.Type_TASK_END_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].end_time = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetEndTime())),
			},
		}

	case events.Type_TASK_OUTPUTS:
		val, err := dynamodbattribute.MarshalList(e.GetOutputs().Value)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB marshal TaskLog Outputs, %v", err)
		}
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].outputs = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				L: val,
			},
		}

	case events.Type_TASK_METADATA:
		val, err := dynamodbattribute.MarshalMap(e.GetMetadata().Value)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB marshal TaskLog Metadata, %v", err)
		}

		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].metadata = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				M: val,
			},
		}

	case events.Type_EXECUTOR_START_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].start_time = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetStartTime())),
			},
		}

	case events.Type_EXECUTOR_END_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].end_time = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetEndTime())),
			},
		}

	case events.Type_EXECUTOR_EXIT_CODE:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].exit_code = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				N: aws.String(strconv.Itoa(int(e.GetExitCode()))),
			},
		}

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
			ExpressionAttributeNames: map[string]*string{
				"#index": aws.String("index"),
			},
			UpdateExpression: aws.String("SET stdout = :stdout, attempt = :attempt, #index = :index"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":stdout": {
					S: aws.String(e.GetStdout()),
				},
				":attempt": {
					N: aws.String(strconv.Itoa(int(e.Attempt))),
				},
				":index": {
					N: aws.String(strconv.Itoa(int(e.Index))),
				},
			},
			ReturnValues: aws.String("UPDATED_NDB"),
		}

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
			ExpressionAttributeNames: map[string]*string{
				"#index": aws.String("index"),
			},
			UpdateExpression: aws.String("SET stderr = :stderr, attempt = :attempt, #index = :index"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":stderr": {
					S: aws.String(e.GetStderr()),
				},
				":attempt": {
					N: aws.String(strconv.Itoa(int(e.Attempt))),
				},
				":index": {
					N: aws.String(strconv.Itoa(int(e.Index))),
				},
			},
			ReturnValues: aws.String("UPDATED_NDB"),
		}
	}

	_, err = db.client.UpdateItemWithContext(ctx, item)
	return err
}

// Close closes the writer.
func (db *DynamoDB) Close() error {
	return nil
}
