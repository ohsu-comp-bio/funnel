package dynamodb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"strconv"
)

// DynamoDB provides handlers for gRPC endpoints
// Data is stored/retrieved from the Amazon DynamoDB NoSQL database.
type DynamoDB struct {
	client         *dynamodb.DynamoDB
	backend        compute.Backend
	partitionKey   string
	partitionValue string
	taskTable      string
	contentsTable  string
	stdoutTable    string
	stderrTable    string
}

// New returns a new instance of DynamoDB, accessing the database at
// the given url, and including the given ServerConfig.
func New(conf config.DynamoDB) (*DynamoDB, error) {
	sess, err := util.NewAWSSession(conf.Key, conf.Secret, conf.Region)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating dynamodb client: %v", err)
	}

	db := &DynamoDB{
		client:         dynamodb.New(sess),
		partitionKey:   "hid",
		partitionValue: "0",
		taskTable:      conf.TableBasename + "-task",
		contentsTable:  conf.TableBasename + "-contents",
		stdoutTable:    conf.TableBasename + "-stdout",
		stderrTable:    conf.TableBasename + "-stderr",
	}

	err = db.createTables()
	return db, err
}

// WithComputeBackend configures the DynamoDB instance to use the given
// compute.Backend. The compute backend is responsible for dispatching tasks to
// schedulers / compute resources with its Submit method.
func (db *DynamoDB) WithComputeBackend(backend compute.Backend) {
	db.backend = backend
}

// CreateTask provides an HTTP/gRPC endpoint for creating a task.
// This is part of the TES implementation.
func (db *DynamoDB) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	log.Debug("CreateTask called", "task", task)

	verr := tes.Validate(task)
	if verr != nil {
		log.Error("Invalid task message", "error", verr)
		return nil, grpc.Errorf(codes.InvalidArgument, verr.Error())
	}

	taskID := util.GenTaskID()
	log := log.WithFields("taskID", taskID)

	task.Id = taskID
	task.State = tes.State_QUEUED

	err := db.createTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}

	err = db.createTaskInputContents(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to write task items to DynamoDB, %v", err)
	}

	err = db.backend.Submit(task)
	if err != nil {
		log.Error("Error submitting task to compute backend", err)

		derr := db.deleteTask(ctx, task.Id)
		if derr != nil {
			err = fmt.Errorf("%v\n%v", err, fmt.Errorf("failed to delete task items from DynamoDB, %v", derr))
		}

		return nil, err
	}

	return &tes.CreateTaskResponse{Id: taskID}, nil
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
			_ = db.getContents(ctx, item)
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
	log := log.WithFields("taskID", req.Id)

	// call GetTask prior to cancel to ensure that the task exists
	t, err := db.GetTask(ctx, &tes.GetTaskRequest{Id: req.Id, View: tes.TaskView_MINIMAL})
	if err != nil {
		return nil, err
	}
	switch t.GetState() {
	case tes.State_COMPLETE, tes.State_ERROR, tes.State_SYSTEM_ERROR:
		err = fmt.Errorf("illegal state transition from %s to %s", t.GetState().String(), tes.State_CANCELED.String())
		log.Error("Cannot cancel task", err)
		return nil, err
	case tes.State_CANCELED:
		log.Info("Canceled task")
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

	log.Info("Canceled task")
	return &tes.CancelTaskResponse{}, nil
}

// GetServiceInfo provides an endpoint for Funnel clients to get information about this server.
func (db *DynamoDB) GetServiceInfo(ctx context.Context, info *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	return &tes.ServiceInfo{}, nil
}
