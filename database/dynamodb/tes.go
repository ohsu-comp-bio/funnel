package dynamodb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

	err = dynamodbattribute.UnmarshalMap(response.Item, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to DynamoDB unmarshal Task, %v", err)
	}

	return task, nil
}

// ListTasks returns a list of taskIDs
func (db *DynamoDB) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	var tasks []*tes.Task
	var query *dynamodb.QueryInput
	pageSize := int64(tes.GetPageSize(req.GetPageSize()))

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

	filterParts := []string{}
	if req.State != tes.Unknown {
		query.ExpressionAttributeNames = map[string]*string{
			"#state": aws.String("state"),
		}
		query.ExpressionAttributeValues[":stateFilter"] = &dynamodb.AttributeValue{
			N: aws.String(strconv.Itoa(int(req.State))),
		}
		filterParts = append(filterParts, "#state = :stateFilter")
	}

	for k, v := range req.GetTags() {
		tmpl := "tags.%s = :%sFilter"
		filterParts = append(filterParts, fmt.Sprintf(tmpl, k, k))
		if v == "" {
			query.ExpressionAttributeValues[fmt.Sprintf(":%sFilter", k)] = &dynamodb.AttributeValue{
				NULL: aws.Bool(true),
			}
		} else {
			query.ExpressionAttributeValues[fmt.Sprintf(":%sFilter", k)] = &dynamodb.AttributeValue{
				S: aws.String(v),
			}
		}
	}

	if len(filterParts) > 0 {
		query.FilterExpression = aws.String(strings.Join(filterParts, " AND "))
	}

	if req.View == tes.View_MINIMAL.String() {
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

	if req.View == tes.View_FULL.String() {
		for _, item := range response.Items {
			// TODO handle errors
			_ = db.getContent(ctx, item)
			_ = db.getExecutorOutput(ctx, item, "stdout", db.stdoutTable)
			_ = db.getExecutorOutput(ctx, item, "stderr", db.stderrTable)
			_ = db.getSystemLogs(ctx, item)
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
