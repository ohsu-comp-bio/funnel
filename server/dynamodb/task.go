package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"strconv"
)

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (db *DynamoDB) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	if err := tes.InitTask(task); err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	err := db.createTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}

	err = db.createTaskInputContent(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}

	err = db.backend.Submit(task)
	if err != nil {
		derr := db.deleteTask(ctx, task.Id)
		if derr != nil {
			err = fmt.Errorf("%v\n%v", err, fmt.Errorf("failed to delete task items from DynamoDB, %v", derr))
		}

		return nil, fmt.Errorf("couldn't submit to compute backend: %s", err)
	}

	return &tes.CreateTaskResponse{Id: task.Id}, nil
}

// GetTask gets a task, which describes a running task
func (db *DynamoDB) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task
	var response *dynamodb.GetItemOutput
	var err error

	switch req.View {
	case tes.TaskView_MINIMAL:
		response, err = db.getMinimalView(ctx, req.Id)
	case tes.TaskView_BASIC:
		response, err = db.getBasicView(ctx, req.Id)
	case tes.TaskView_FULL:
		response, err = db.getFullView(ctx, req.Id)
	}
	if err != nil {
		return nil, err
	}

	if response.Item == nil {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: taskID: %s", errNotFound.Error(), req.Id))
	}

	err = dynamodbattribute.UnmarshalMap(response.Item, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to DynamoDB unmarshal Task, %v", err)
	}

	return task, nil
}

// ListTasks returns a list of taskIDs
func (db *DynamoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	var tasks []*tes.Task
	var pageSize int64 = 256
	var query *dynamodb.QueryInput

	if req.PageSize != 0 {
		pageSize = int64(req.GetPageSize())
		if pageSize > 2048 {
			pageSize = 2048
		}
		if pageSize < 50 {
			pageSize = 50
		}
	}

	query = &dynamodb.QueryInput{
		TableName:              aws.String(db.taskTable),
		Limit:                  aws.Int64(pageSize),
		ScanIndexForward:       aws.Bool(false),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :v1", db.partitionKey)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(db.partitionValue),
			},
		},
	}

	if req.View == tes.TaskView_MINIMAL {
		query.ExpressionAttributeNames = map[string]*string{
			"#state": aws.String("state"),
		}
		query.ProjectionExpression = aws.String("id, #state")
	}

	if req.PageToken != "" {
		query.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(req.PageToken),
			},
		}
	}

	response, err := db.client.QueryWithContext(ctx, query)

	if err != nil {
		return nil, err
	}

	if req.View == tes.TaskView_FULL {
		for _, item := range response.Items {
			// TODO handle errors
			_ = db.getContent(ctx, item)
			_ = db.getExecutorOutput(ctx, item, "stdout", db.stdoutTable)
			_ = db.getExecutorOutput(ctx, item, "stderr", db.stderrTable)
		}
	}

	err = dynamodbattribute.UnmarshalListOfMaps(response.Items, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to DynamoDB unmarshal Tasks, %v", err)
	}

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	if response.LastEvaluatedKey != nil {
		out.NextPageToken = *response.LastEvaluatedKey["id"].S
	}

	return &out, nil
}

// CancelTask cancels a task
func (db *DynamoDB) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {

	// call GetTask prior to cancel to ensure that the task exists
	t, err := db.GetTask(ctx, &tes.GetTaskRequest{Id: req.Id, View: tes.TaskView_MINIMAL})
	if err != nil {
		return nil, err
	}
	switch t.GetState() {
	case tes.State_COMPLETE, tes.State_EXECUTOR_ERROR, tes.State_SYSTEM_ERROR:
		err = fmt.Errorf("illegal state transition from %s to %s", t.GetState().String(), tes.State_CANCELED.String())
		return nil, fmt.Errorf("cannot cancel task: %s", err)
	case tes.State_CANCELED:
		return &tes.CancelTaskResponse{}, nil
	}

	item := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(req.Id),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#state": aws.String("state"),
		},
		UpdateExpression: aws.String("SET #state = :to"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":to": {
				N: aws.String(strconv.Itoa(int(tes.State_CANCELED))),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err = db.client.UpdateItemWithContext(ctx, item)
	if err != nil {
		return nil, err
	}

	return &tes.CancelTaskResponse{}, nil
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
func (db *DynamoDB) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{}, nil
}
