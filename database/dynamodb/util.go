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
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
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

	table = &dynamodb.CreateTableInput{
		TableName: aws.String(db.syslogsTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("attempt"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("attempt"),
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
	return db.waitForTables()
}

func (db *DynamoDB) tableIsAlive(ctx context.Context, name string) error {
	ticker := time.NewTicker(time.Millisecond * 500).C
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker:
			r, err := db.client.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(name),
			})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
						continue
					}
				}
				return err
			}
			if *r.Table.TableStatus == "ACTIVE" {
				return nil
			}
		}
	}
}

func (db *DynamoDB) waitForTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	if err := db.tableIsAlive(ctx, db.taskTable); err != nil {
		return err
	}
	if err := db.tableIsAlive(ctx, db.contentTable); err != nil {
		return err
	}
	if err := db.tableIsAlive(ctx, db.stdoutTable); err != nil {
		return err
	}
	if err := db.tableIsAlive(ctx, db.stderrTable); err != nil {
		return err
	}
	if err := db.tableIsAlive(ctx, db.syslogsTable); err != nil {
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

	av["version"] = &dynamodb.AttributeValue{
		S: aws.String(strconv.FormatInt(time.Now().UnixNano(), 10)),
	}

	av["owner"] = &dynamodb.AttributeValue{
		S: aws.String(server.GetUsername(ctx)),
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
	if err != nil {
		return fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}

	err = db.createTaskInputContent(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}
	return nil
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

func (db *DynamoDB) taskKey(id string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		db.partitionKey: {
			S: aws.String(db.partitionValue),
		},
		"id": {
			S: aws.String(id),
		},
	}
}

func (db *DynamoDB) getMinimalView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	item := &dynamodb.GetItemInput{
		TableName: aws.String(db.taskTable),
		Key:       db.taskKey(id),
		ExpressionAttributeNames: map[string]*string{
			"#owner": aws.String("owner"),
			"#state": aws.String("state"),
		},
		ProjectionExpression: aws.String("id, #owner, #state"),
	}
	return db.client.GetItemWithContext(ctx, item)
}

func (db *DynamoDB) getBasicView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	item := &dynamodb.GetItemInput{
		TableName: aws.String(db.taskTable),
		Key:       db.taskKey(id),
	}
	return db.client.GetItemWithContext(ctx, item)
}

func (db *DynamoDB) getFullView(ctx context.Context, id string) (*dynamodb.GetItemOutput, error) {
	resp, err := db.getBasicView(ctx, id)
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

	err = db.getSystemLogs(ctx, resp.Item)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve system logs: %v", err)
	}

	return resp, nil
}
func (db *DynamoDB) queryInput(table string, id *dynamodb.AttributeValue, limit int64) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:              aws.String(table),
		Limit:                  aws.Int64(limit),
		ScanIndexForward:       aws.Bool(false),
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("id = :v1"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": id,
		},
	}
}

func (db *DynamoDB) getContent(ctx context.Context, in map[string]*dynamodb.AttributeValue) error {
	return db.client.QueryPagesWithContext(
		ctx,
		db.queryInput(db.contentTable, in["id"], 10),
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				i, _ := strconv.ParseInt(*item["index"].N, 10, 64)
				in["inputs"].L[i].M["content"] = item["content"]
			}
			return page.LastEvaluatedKey != nil
		},
	)
}

func (db *DynamoDB) getExecutorOutput(ctx context.Context, in map[string]*dynamodb.AttributeValue, val string, table string) error {
	return db.client.QueryPagesWithContext(
		ctx,
		db.queryInput(table, in["id"], 10),
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
			return page.LastEvaluatedKey != nil
		},
	)
}

func (db *DynamoDB) getSystemLogs(ctx context.Context, in map[string]*dynamodb.AttributeValue) error {
	return db.client.QueryPagesWithContext(
		ctx,
		db.queryInput(db.syslogsTable, in["id"], 50),
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				i, _ := strconv.ParseInt(*item["attempt"].N, 10, 64)
				in["logs"].L[i].M["system_logs"] = item["system_logs"]
			}
			return page.LastEvaluatedKey != nil
		},
	)
}

func isAccessible(ctx context.Context, response *dynamodb.GetItemOutput) bool {
	taskOwner := ""
	if attrValue, ok := response.Item["owner"]; ok {
		taskOwner = *attrValue.S
	}
	return server.GetUser(ctx).IsAccessible(taskOwner)
}
