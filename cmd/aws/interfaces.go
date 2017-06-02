package aws

import (
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Batch provides a simple interface to the AWS Batch API used by Funnel.
// This is useful for testing.
type Batch interface {
	SubmitJob(*batch.SubmitJobInput) (*batch.SubmitJobOutput, error)
	DescribeJobs(*batch.DescribeJobsInput) (*batch.DescribeJobsOutput, error)
	TerminateJob(*batch.TerminateJobInput) (*batch.TerminateJobOutput, error)
	ListJobs(*batch.ListJobsInput) (*batch.ListJobsOutput, error)
	CreateComputeEnvironment(*batch.CreateComputeEnvironmentInput) (*batch.CreateComputeEnvironmentOutput, error)
	CreateJobQueue(*batch.CreateJobQueueInput) (*batch.CreateJobQueueOutput, error)
	RegisterJobDefinition(*batch.RegisterJobDefinitionInput) (*batch.RegisterJobDefinitionOutput, error)
}

// CloudWatchLogs provides a simple interface to the AWS APIs used by Funnel.
// This is useful for testing.
type CloudWatchLogs interface {
	GetLogEvents(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error)
}
