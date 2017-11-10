package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strconv"
)

func checkCreateErr(err error) error {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeResourceInUseException:
			return nil
		}
	}
	return err
}

func (db *DynamoDB) createTables() error {
	var table *dynamodb.CreateTableInput
	var err error

	table = &dynamodb.CreateTableInput{
		TableName: aws.String(db.taskTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(db.partitionKey),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(db.partitionKey),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	}
	_, err = db.client.CreateTable(table)
	if checkCreateErr(err) != nil {
		return err
	}

	table = &dynamodb.CreateTableInput{
		TableName: aws.String(db.contentTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("index"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("index"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	}
	_, err = db.client.CreateTable(table)
	if checkCreateErr(err) != nil {
		return err
	}

	table = &dynamodb.CreateTableInput{
		TableName: aws.String(db.stdoutTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("attempt_index"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("attempt_index"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	}
	_, err = db.client.CreateTable(table)
	if checkCreateErr(err) != nil {
		return err
	}

	table = &dynamodb.CreateTableInput{
		TableName: aws.String(db.stderrTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("attempt_index"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("attempt_index"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	}
	_, err = db.client.CreateTable(table)
	if checkCreateErr(err) != nil {
		return err
	}
	return nil
}

func (db *DynamoDB) createTask(ctx context.Context, task *tes.Task) error {
	taskBasic := task.GetBasicView()
	av, err := dynamodbattribute.MarshalMap(taskBasic)
	if err != nil {
		return fmt.Errorf("failed to DynamoDB marshal Task, %v", err)
	}

	av[db.partitionKey] = &dynamodb.AttributeValue{
		S: aws.String(db.partitionValue),
	}

	av["created_at"] = &dynamodb.AttributeValue{
		S: aws.String(ptypes.TimestampString(ptypes.TimestampNow())),
	}

	// Add nil fields to make updates easier
	av["logs"] = &dynamodb.AttributeValue{
		L: []*dynamodb.AttributeValue{
			{
				M: map[string]*dynamodb.AttributeValue{
					"logs": {
						L: []*dynamodb.AttributeValue{},
					},
					"outputs": {
						L: []*dynamodb.AttributeValue{},
					},
				},
			},
		},
	}

	item := &dynamodb.PutItemInput{
		TableName: aws.String(db.taskTable),
		Item:      av,
	}

	_, err = db.client.PutItemWithContext(ctx, item)
	return err
}

func (db *DynamoDB) createTaskInputContent(ctx context.Context, task *tes.Task) error {
	av := make(map[string]*dynamodb.AttributeValue)

	for i, v := range task.Inputs {
		if v.Content != "" {
			av["id"] = &dynamodb.AttributeValue{
				S: aws.String(task.Id),
			}
			av["index"] = &dynamodb.AttributeValue{
				N: aws.String(strconv.Itoa(i)),
			}
			av["content"] = &dynamodb.AttributeValue{
				S: aws.String(v.Content),
			}
			av["created_at"] = &dynamodb.AttributeValue{
				S: aws.String(ptypes.TimestampString(ptypes.TimestampNow())),
			}

			item := &dynamodb.PutItemInput{
				TableName: aws.String(db.contentTable),
				Item:      av,
			}

			_, err := db.client.PutItemWithContext(ctx, item)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DynamoDB) deleteTask(ctx context.Context, id string) error {
	var item *dynamodb.DeleteItemInput
	var err error

	item = &dynamodb.DeleteItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
	}
	_, err = db.client.DeleteItemWithContext(ctx, item)
	if err != nil {
		return err
	}

	query := &dynamodb.QueryInput{
		TableName:              aws.String(db.contentTable),
		Limit:                  aws.Int64(10),
		ScanIndexForward:       aws.Bool(false),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("id = :v1"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(id),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#index": aws.String("index"),
		},
		ProjectionExpression: aws.String("id, #index"),
	}

	err = db.client.QueryPagesWithContext(
		ctx,
		query,
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, res := range page.Items {
				item = &dynamodb.DeleteItemInput{
					TableName: aws.String(db.contentTable),
					Key: map[string]*dynamodb.AttributeValue{
						"id":    res["id"],
						"index": res["index"],
					},
				}
				// TODO handle error
				db.client.DeleteItem(item)
			}
			if page.LastEvaluatedKey == nil {
				return false
			}
			return true
		})

	if err != nil {
		return err
	}

	return nil
}

func (db *DynamoDB) getMinimalView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	item := &dynamodb.GetItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#state": aws.String("state"),
		},
		ProjectionExpression: aws.String("id, #state"),
	}
	return db.client.GetItemWithContext(ctx, item)
}

func (db *DynamoDB) getBasicView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	item := &dynamodb.GetItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
	}
	return db.client.GetItemWithContext(ctx, item)
}

func (db *DynamoDB) getFullView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	item := &dynamodb.GetItemInput{
		TableName: aws.String(db.taskTable),
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {
				S: aws.String(db.partitionValue),
			},
			"id": {
				S: aws.String(id),
			},
		},
	}

	resp, err := db.client.GetItemWithContext(ctx, item)
	if err != nil || resp.Item == nil {
		return resp, err
	}

	err = db.getContent(ctx, resp.Item)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve input content: %v", err)
	}

	err = db.getExecutorOutput(ctx, resp.Item, "stdout", db.stdoutTable)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve task executor log stdout: %v", err)
	}

	err = db.getExecutorOutput(ctx, resp.Item, "stderr", db.stderrTable)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve task executor log stderr: %v", err)
	}

	return resp, nil
}

func (db *DynamoDB) getContent(ctx context.Context, in map[string]*dynamodb.AttributeValue) error {
	query := &dynamodb.QueryInput{
		TableName:              aws.String(db.contentTable),
		Limit:                  aws.Int64(10),
		ScanIndexForward:       aws.Bool(false),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("id = :v1"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": in["id"],
		},
	}

	err := db.client.QueryPagesWithContext(
		ctx,
		query,
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				i, _ := strconv.ParseInt(*item["index"].N, 10, 64)
				in["inputs"].L[i].M["content"] = item["content"]
			}
			if page.LastEvaluatedKey == nil {
				return false
			}
			return true
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *DynamoDB) getExecutorOutput(ctx context.Context, in map[string]*dynamodb.AttributeValue, val string, table string) error {
	query := &dynamodb.QueryInput{
		TableName:              aws.String(table),
		Limit:                  aws.Int64(10),
		ScanIndexForward:       aws.Bool(false),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("id = :v1"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": in["id"],
		},
	}

	err := db.client.QueryPagesWithContext(
		ctx,
		query,
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				i, _ := strconv.ParseInt(*item["index"].N, 10, 64)
				a, _ := strconv.ParseInt(*item["attempt"].N, 10, 64)
				if out, ok := item[val]; ok {
					in["logs"].L[a].M["logs"].L[i].M[val] = &dynamodb.AttributeValue{
						S: aws.String(*out.S),
					}
				}
			}
			if page.LastEvaluatedKey == nil {
				return false
			}
			return true
		},
	)
	if err != nil {
		return err
	}
	return nil
}
