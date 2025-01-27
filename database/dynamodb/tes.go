package dynamodb

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// GetTask gets a task, which describes a running task
func (db *DynamoDB) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task
	var response *dynamodb.GetItemOutput
	var err error

	switch req.View {
	case tes.View_MINIMAL.String():
		response, err = db.getMinimalView(ctx, req.Id)
	case tes.View_BASIC.String():
		response, err = db.getBasicView(ctx, req.Id)
	case tes.View_FULL.String():
		response, err = db.getFullView(ctx, req.Id)
	}
	if err != nil {
		return nil, err
	}

	if response.Item == nil {
		return nil, tes.ErrNotFound
	}

	taskOwner := ""
	if attrValue, ok := response.Item["owner"]; ok {
		taskOwner = *attrValue.S
	}
	if !server.GetUser(ctx).IsAccessible(taskOwner) {
		return nil, tes.ErrNotPermitted
	}

	err = dynamodbattribute.UnmarshalMap(response.Item, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to DynamoDB unmarshal Task, %v", err)
	}

	return task, nil
}

// ListTasks returns a list of taskIDs
func (db *DynamoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	exp := expression.NewBuilder().
		WithKeyCondition(expression.Key(db.partitionKey).Equal(expression.Value(db.partitionValue)))

	if userInfo := server.GetUser(ctx); !userInfo.CanSeeAllTasks() {
		exp = exp.WithFilter(expression.Name("owner").Equal(expression.Value(userInfo.Username)))
	}

	if req.State != tes.Unknown {
		numStr := strconv.Itoa(int(req.State))
		exp = exp.WithFilter(expression.Name("state").Equal(expression.Value(numStr)))
	}

	for k, v := range req.GetTags() {
		var fieldValue expression.ValueBuilder
		if v == "" {
			fieldValue = expression.Value(expression.Null)
		} else {
			fieldValue = expression.Value(v)
		}
		exp = exp.WithFilter(expression.Name("tags." + k).Equal(fieldValue))
	}

	if req.View == tes.View_MINIMAL.String() {
		exp = exp.WithProjection(expression.NamesList(expression.Name("id"), expression.Name("state")))
	}

	eb, err := exp.Build()
	if err != nil {
		return nil, err
	}

	pageSize := int64(tes.GetPageSize(req.GetPageSize()))
	query := &dynamodb.QueryInput{
		TableName:                 aws.String(db.taskTable),
		Limit:                     aws.Int64(pageSize),
		ScanIndexForward:          aws.Bool(false),
		ConsistentRead:            aws.Bool(true),
		KeyConditionExpression:    eb.KeyCondition(),
		ExpressionAttributeNames:  eb.Names(),
		ExpressionAttributeValues: eb.Values(),
		FilterExpression:          eb.Filter(),
		ProjectionExpression:      eb.Projection(),
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

	if req.View == tes.View_FULL.String() {
		for _, item := range response.Items {
			// TODO handle errors
			_ = db.getContent(ctx, item)
			_ = db.getExecutorOutput(ctx, item, "stdout", db.stdoutTable)
			_ = db.getExecutorOutput(ctx, item, "stderr", db.stderrTable)
			_ = db.getSystemLogs(ctx, item)
		}
	}

	var tasks []*tes.Task
	err = dynamodbattribute.UnmarshalListOfMaps(response.Items, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to DynamoDB unmarshal Tasks, %v", err)
	}

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	if len(tasks) > 0 && response.LastEvaluatedKey != nil {
		out.NextPageToken = *response.LastEvaluatedKey["id"].S
	}

	return &out, nil
}
