package events

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"strconv"
)

// DynamoDBEventWriter is a type which writes Events to DynamoDB.
type DynamoDBEventWriter struct {
	client         *dynamodb.DynamoDB
	partitionKey   string
	partitionValue string
	taskTable      string
	contentsTable  string
	stdoutTable    string
	stderrTable    string
}

// NewDynamoDBEventWriter returns a new DynamoDBEventWriter instance.
func NewDynamoDBEventWriter(conf config.DynamoDB) (*DynamoDBEventWriter, error) {
	sess, err := util.NewAWSSession(conf.Key, conf.Secret, conf.Region)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating dynamodb client: %v", err)
	}

	return &DynamoDBEventWriter{
		client:         dynamodb.New(sess),
		partitionKey:   "hid",
		partitionValue: "0",
		taskTable:      conf.TableBasename + "-task",
		contentsTable:  conf.TableBasename + "-contents",
		stdoutTable:    conf.TableBasename + "-stdout",
		stderrTable:    conf.TableBasename + "-stderr",
	}, nil
}

// Write writes an event to DynamoDB.
func (ew *DynamoDBEventWriter) Write(e *Event) error {
	item := &dynamodb.UpdateItemInput{
		TableName: aws.String(ew.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			ew.partitionKey: {
				S: aws.String(ew.partitionValue),
			},
			"id": {
				S: aws.String(e.Id),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	// create the log structure for the attempt if it doesnt already exist
	attemptItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(ew.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			ew.partitionKey: {
				S: aws.String(ew.partitionValue),
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
		ReturnValues: aws.String("UPDATED_NEW"),
	}
	_, err := ew.client.UpdateItem(attemptItem)
	if err != nil {
		return err
	}

	// create the log structure for the executor if it doesnt already exist
	indexItem := &dynamodb.UpdateItemInput{
		TableName: aws.String(ew.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			ew.partitionKey: {
				S: aws.String(ew.partitionValue),
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
		ReturnValues: aws.String("UPDATED_NEW"),
	}
	_, err = ew.client.UpdateItem(indexItem)
	if err != nil {
		return err
	}

	switch e.Type {
	case Type_TASK_STATE:
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

		case tes.State_ERROR:
			item.ConditionExpression = aws.String("#state IN (:i, :r)")
			item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
				":to": {
					N: aws.String(strconv.Itoa(int(tes.State_ERROR))),
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

	case Type_TASK_START_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].start_time = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetStartTime())),
			},
		}

	case Type_TASK_END_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].end_time = :c", e.Attempt))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetEndTime())),
			},
		}

	case Type_TASK_OUTPUTS:
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

	case Type_TASK_METADATA:
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

	case Type_EXECUTOR_START_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].start_time = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetStartTime())),
			},
		}

	case Type_EXECUTOR_END_TIME:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].end_time = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(ptypes.TimestampString(e.GetEndTime())),
			},
		}

	case Type_EXECUTOR_EXIT_CODE:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].exit_code = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				N: aws.String(strconv.Itoa(int(e.GetExitCode()))),
			},
		}

	case Type_EXECUTOR_HOST_IP:
		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].host_ip = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				S: aws.String(e.GetHostIp()),
			},
		}

	case Type_EXECUTOR_PORTS:
		val, err := dynamodbattribute.MarshalMap(e.GetPorts().Value)
		if err != nil {
			return fmt.Errorf("failed to DynamoDB marshal ExecutorLog Ports, %v", err)
		}

		item.UpdateExpression = aws.String(fmt.Sprintf("SET logs[%v].logs[%v].ports = :c", e.Attempt, e.Index))
		item.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":c": {
				M: val,
			},
		}

	case Type_EXECUTOR_STDOUT:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(ew.stdoutTable),
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
			UpdateExpression: aws.String("SET stdout = list_append(if_not_exists(stdout, :e), :stdout), attempt = :attempt, #index = :index"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":e": {
					L: []*dynamodb.AttributeValue{},
				},
				":stdout": {
					L: []*dynamodb.AttributeValue{
						{
							M: map[string]*dynamodb.AttributeValue{
								"value": {
									S: aws.String(e.GetStdout()),
								},
								"timestamp": {
									S: aws.String(ptypes.TimestampString(ptypes.TimestampNow())),
								},
							},
						},
					},
				},
				":attempt": {
					N: aws.String(strconv.Itoa(int(e.Attempt))),
				},
				":index": {
					N: aws.String(strconv.Itoa(int(e.Index))),
				},
			},
			ReturnValues: aws.String("UPDATED_NEW"),
		}

	case Type_EXECUTOR_STDERR:
		item = &dynamodb.UpdateItemInput{
			TableName: aws.String(ew.stderrTable),
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
			UpdateExpression: aws.String("SET stderr = list_append(if_not_exists(stderr, :e), :stderr), attempt = :attempt, #index = :index"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":e": {
					L: []*dynamodb.AttributeValue{},
				},
				":stderr": {
					L: []*dynamodb.AttributeValue{
						{
							M: map[string]*dynamodb.AttributeValue{
								"value": {
									S: aws.String(e.GetStderr()),
								},
								"timestamp": {
									S: aws.String(ptypes.TimestampString(ptypes.TimestampNow())),
								},
							},
						},
					},
				},
				":attempt": {
					N: aws.String(strconv.Itoa(int(e.Attempt))),
				},
				":index": {
					N: aws.String(strconv.Itoa(int(e.Index))),
				},
			},
			ReturnValues: aws.String("UPDATED_NEW"),
		}
	}

	_, err = ew.client.UpdateItem(item)
	return err
}
