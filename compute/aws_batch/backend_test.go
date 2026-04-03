package aws_batch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// mockBatchClient is a minimal mock of the BatchAPI interface for testing.
type mockBatchClient struct {
	submitJobFunc func(*batch.SubmitJobInput) (*batch.SubmitJobOutput, error)
}

func (m *mockBatchClient) SubmitJob(input *batch.SubmitJobInput) (*batch.SubmitJobOutput, error) {
	if m.submitJobFunc != nil {
		return m.submitJobFunc(input)
	}
	return &batch.SubmitJobOutput{JobId: aws.String("test-job-id"), JobName: input.JobName}, nil
}

func (m *mockBatchClient) SubmitJobWithContext(_ aws.Context, input *batch.SubmitJobInput, _ ...request.Option) (*batch.SubmitJobOutput, error) {
	return m.SubmitJob(input)
}

func (m *mockBatchClient) SubmitJobRequest(input *batch.SubmitJobInput) (*request.Request, *batch.SubmitJobOutput) {
	return nil, nil
}

// Implement all other required methods of BatchAPI with no-ops.
func (m *mockBatchClient) CancelJob(*batch.CancelJobInput) (*batch.CancelJobOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CancelJobWithContext(aws.Context, *batch.CancelJobInput, ...request.Option) (*batch.CancelJobOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CancelJobRequest(*batch.CancelJobInput) (*request.Request, *batch.CancelJobOutput) {
	return nil, nil
}
func (m *mockBatchClient) CreateComputeEnvironment(*batch.CreateComputeEnvironmentInput) (*batch.CreateComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateComputeEnvironmentWithContext(aws.Context, *batch.CreateComputeEnvironmentInput, ...request.Option) (*batch.CreateComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateComputeEnvironmentRequest(*batch.CreateComputeEnvironmentInput) (*request.Request, *batch.CreateComputeEnvironmentOutput) {
	return nil, nil
}
func (m *mockBatchClient) CreateJobQueue(*batch.CreateJobQueueInput) (*batch.CreateJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateJobQueueWithContext(aws.Context, *batch.CreateJobQueueInput, ...request.Option) (*batch.CreateJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateJobQueueRequest(*batch.CreateJobQueueInput) (*request.Request, *batch.CreateJobQueueOutput) {
	return nil, nil
}
func (m *mockBatchClient) CreateSchedulingPolicy(*batch.CreateSchedulingPolicyInput) (*batch.CreateSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateSchedulingPolicyWithContext(aws.Context, *batch.CreateSchedulingPolicyInput, ...request.Option) (*batch.CreateSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) CreateSchedulingPolicyRequest(*batch.CreateSchedulingPolicyInput) (*request.Request, *batch.CreateSchedulingPolicyOutput) {
	return nil, nil
}
func (m *mockBatchClient) DeleteComputeEnvironment(*batch.DeleteComputeEnvironmentInput) (*batch.DeleteComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteComputeEnvironmentWithContext(aws.Context, *batch.DeleteComputeEnvironmentInput, ...request.Option) (*batch.DeleteComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteComputeEnvironmentRequest(*batch.DeleteComputeEnvironmentInput) (*request.Request, *batch.DeleteComputeEnvironmentOutput) {
	return nil, nil
}
func (m *mockBatchClient) DeleteJobQueue(*batch.DeleteJobQueueInput) (*batch.DeleteJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteJobQueueWithContext(aws.Context, *batch.DeleteJobQueueInput, ...request.Option) (*batch.DeleteJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteJobQueueRequest(*batch.DeleteJobQueueInput) (*request.Request, *batch.DeleteJobQueueOutput) {
	return nil, nil
}
func (m *mockBatchClient) DeleteSchedulingPolicy(*batch.DeleteSchedulingPolicyInput) (*batch.DeleteSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteSchedulingPolicyWithContext(aws.Context, *batch.DeleteSchedulingPolicyInput, ...request.Option) (*batch.DeleteSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeleteSchedulingPolicyRequest(*batch.DeleteSchedulingPolicyInput) (*request.Request, *batch.DeleteSchedulingPolicyOutput) {
	return nil, nil
}
func (m *mockBatchClient) DeregisterJobDefinition(*batch.DeregisterJobDefinitionInput) (*batch.DeregisterJobDefinitionOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeregisterJobDefinitionWithContext(aws.Context, *batch.DeregisterJobDefinitionInput, ...request.Option) (*batch.DeregisterJobDefinitionOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DeregisterJobDefinitionRequest(*batch.DeregisterJobDefinitionInput) (*request.Request, *batch.DeregisterJobDefinitionOutput) {
	return nil, nil
}
func (m *mockBatchClient) DescribeComputeEnvironments(*batch.DescribeComputeEnvironmentsInput) (*batch.DescribeComputeEnvironmentsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeComputeEnvironmentsWithContext(aws.Context, *batch.DescribeComputeEnvironmentsInput, ...request.Option) (*batch.DescribeComputeEnvironmentsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeComputeEnvironmentsRequest(*batch.DescribeComputeEnvironmentsInput) (*request.Request, *batch.DescribeComputeEnvironmentsOutput) {
	return nil, nil
}
func (m *mockBatchClient) DescribeComputeEnvironmentsPages(*batch.DescribeComputeEnvironmentsInput, func(*batch.DescribeComputeEnvironmentsOutput, bool) bool) error {
	return nil
}
func (m *mockBatchClient) DescribeComputeEnvironmentsPagesWithContext(aws.Context, *batch.DescribeComputeEnvironmentsInput, func(*batch.DescribeComputeEnvironmentsOutput, bool) bool, ...request.Option) error {
	return nil
}
func (m *mockBatchClient) DescribeJobDefinitions(*batch.DescribeJobDefinitionsInput) (*batch.DescribeJobDefinitionsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobDefinitionsWithContext(aws.Context, *batch.DescribeJobDefinitionsInput, ...request.Option) (*batch.DescribeJobDefinitionsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobDefinitionsRequest(*batch.DescribeJobDefinitionsInput) (*request.Request, *batch.DescribeJobDefinitionsOutput) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobDefinitionsPages(*batch.DescribeJobDefinitionsInput, func(*batch.DescribeJobDefinitionsOutput, bool) bool) error {
	return nil
}
func (m *mockBatchClient) DescribeJobDefinitionsPagesWithContext(aws.Context, *batch.DescribeJobDefinitionsInput, func(*batch.DescribeJobDefinitionsOutput, bool) bool, ...request.Option) error {
	return nil
}
func (m *mockBatchClient) DescribeJobQueues(*batch.DescribeJobQueuesInput) (*batch.DescribeJobQueuesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobQueuesWithContext(aws.Context, *batch.DescribeJobQueuesInput, ...request.Option) (*batch.DescribeJobQueuesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobQueuesRequest(*batch.DescribeJobQueuesInput) (*request.Request, *batch.DescribeJobQueuesOutput) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobQueuesPages(*batch.DescribeJobQueuesInput, func(*batch.DescribeJobQueuesOutput, bool) bool) error {
	return nil
}
func (m *mockBatchClient) DescribeJobQueuesPagesWithContext(aws.Context, *batch.DescribeJobQueuesInput, func(*batch.DescribeJobQueuesOutput, bool) bool, ...request.Option) error {
	return nil
}
func (m *mockBatchClient) DescribeJobs(*batch.DescribeJobsInput) (*batch.DescribeJobsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobsWithContext(aws.Context, *batch.DescribeJobsInput, ...request.Option) (*batch.DescribeJobsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeJobsRequest(*batch.DescribeJobsInput) (*request.Request, *batch.DescribeJobsOutput) {
	return nil, nil
}
func (m *mockBatchClient) DescribeSchedulingPolicies(*batch.DescribeSchedulingPoliciesInput) (*batch.DescribeSchedulingPoliciesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeSchedulingPoliciesWithContext(aws.Context, *batch.DescribeSchedulingPoliciesInput, ...request.Option) (*batch.DescribeSchedulingPoliciesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) DescribeSchedulingPoliciesRequest(*batch.DescribeSchedulingPoliciesInput) (*request.Request, *batch.DescribeSchedulingPoliciesOutput) {
	return nil, nil
}
func (m *mockBatchClient) GetJobQueueSnapshot(*batch.GetJobQueueSnapshotInput) (*batch.GetJobQueueSnapshotOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) GetJobQueueSnapshotWithContext(aws.Context, *batch.GetJobQueueSnapshotInput, ...request.Option) (*batch.GetJobQueueSnapshotOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) GetJobQueueSnapshotRequest(*batch.GetJobQueueSnapshotInput) (*request.Request, *batch.GetJobQueueSnapshotOutput) {
	return nil, nil
}
func (m *mockBatchClient) ListJobs(*batch.ListJobsInput) (*batch.ListJobsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListJobsWithContext(aws.Context, *batch.ListJobsInput, ...request.Option) (*batch.ListJobsOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListJobsRequest(*batch.ListJobsInput) (*request.Request, *batch.ListJobsOutput) {
	return nil, nil
}
func (m *mockBatchClient) ListJobsPages(*batch.ListJobsInput, func(*batch.ListJobsOutput, bool) bool) error {
	return nil
}
func (m *mockBatchClient) ListJobsPagesWithContext(aws.Context, *batch.ListJobsInput, func(*batch.ListJobsOutput, bool) bool, ...request.Option) error {
	return nil
}
func (m *mockBatchClient) ListSchedulingPolicies(*batch.ListSchedulingPoliciesInput) (*batch.ListSchedulingPoliciesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListSchedulingPoliciesWithContext(aws.Context, *batch.ListSchedulingPoliciesInput, ...request.Option) (*batch.ListSchedulingPoliciesOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListSchedulingPoliciesRequest(*batch.ListSchedulingPoliciesInput) (*request.Request, *batch.ListSchedulingPoliciesOutput) {
	return nil, nil
}
func (m *mockBatchClient) ListSchedulingPoliciesPages(*batch.ListSchedulingPoliciesInput, func(*batch.ListSchedulingPoliciesOutput, bool) bool) error {
	return nil
}
func (m *mockBatchClient) ListSchedulingPoliciesPagesWithContext(aws.Context, *batch.ListSchedulingPoliciesInput, func(*batch.ListSchedulingPoliciesOutput, bool) bool, ...request.Option) error {
	return nil
}
func (m *mockBatchClient) ListTagsForResource(*batch.ListTagsForResourceInput) (*batch.ListTagsForResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListTagsForResourceWithContext(aws.Context, *batch.ListTagsForResourceInput, ...request.Option) (*batch.ListTagsForResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) ListTagsForResourceRequest(*batch.ListTagsForResourceInput) (*request.Request, *batch.ListTagsForResourceOutput) {
	return nil, nil
}
func (m *mockBatchClient) RegisterJobDefinition(*batch.RegisterJobDefinitionInput) (*batch.RegisterJobDefinitionOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) RegisterJobDefinitionWithContext(aws.Context, *batch.RegisterJobDefinitionInput, ...request.Option) (*batch.RegisterJobDefinitionOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) RegisterJobDefinitionRequest(*batch.RegisterJobDefinitionInput) (*request.Request, *batch.RegisterJobDefinitionOutput) {
	return nil, nil
}
func (m *mockBatchClient) TagResource(*batch.TagResourceInput) (*batch.TagResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) TagResourceWithContext(aws.Context, *batch.TagResourceInput, ...request.Option) (*batch.TagResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) TagResourceRequest(*batch.TagResourceInput) (*request.Request, *batch.TagResourceOutput) {
	return nil, nil
}
func (m *mockBatchClient) TerminateJob(*batch.TerminateJobInput) (*batch.TerminateJobOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) TerminateJobWithContext(aws.Context, *batch.TerminateJobInput, ...request.Option) (*batch.TerminateJobOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) TerminateJobRequest(*batch.TerminateJobInput) (*request.Request, *batch.TerminateJobOutput) {
	return nil, nil
}
func (m *mockBatchClient) UntagResource(*batch.UntagResourceInput) (*batch.UntagResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UntagResourceWithContext(aws.Context, *batch.UntagResourceInput, ...request.Option) (*batch.UntagResourceOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UntagResourceRequest(*batch.UntagResourceInput) (*request.Request, *batch.UntagResourceOutput) {
	return nil, nil
}
func (m *mockBatchClient) UpdateComputeEnvironment(*batch.UpdateComputeEnvironmentInput) (*batch.UpdateComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateComputeEnvironmentWithContext(aws.Context, *batch.UpdateComputeEnvironmentInput, ...request.Option) (*batch.UpdateComputeEnvironmentOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateComputeEnvironmentRequest(*batch.UpdateComputeEnvironmentInput) (*request.Request, *batch.UpdateComputeEnvironmentOutput) {
	return nil, nil
}
func (m *mockBatchClient) UpdateJobQueue(*batch.UpdateJobQueueInput) (*batch.UpdateJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateJobQueueWithContext(aws.Context, *batch.UpdateJobQueueInput, ...request.Option) (*batch.UpdateJobQueueOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateJobQueueRequest(*batch.UpdateJobQueueInput) (*request.Request, *batch.UpdateJobQueueOutput) {
	return nil, nil
}
func (m *mockBatchClient) UpdateSchedulingPolicy(*batch.UpdateSchedulingPolicyInput) (*batch.UpdateSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateSchedulingPolicyWithContext(aws.Context, *batch.UpdateSchedulingPolicyInput, ...request.Option) (*batch.UpdateSchedulingPolicyOutput, error) {
	return nil, nil
}
func (m *mockBatchClient) UpdateSchedulingPolicyRequest(*batch.UpdateSchedulingPolicyInput) (*request.Request, *batch.UpdateSchedulingPolicyOutput) {
	return nil, nil
}

// noopEventWriter satisfies the events.Writer interface with no-op implementations.
type noopEventWriter struct{}

func (n *noopEventWriter) WriteEvent(_ context.Context, _ *events.Event) error { return nil }
func (n *noopEventWriter) Close()                                               {}

func TestCreateJobc(t *testing.T) {
	log := logger.NewLogger("test", logger.DefaultConfig())

	var capturedInput *batch.SubmitJobInput
	mockClient := &mockBatchClient{
		submitJobFunc: func(input *batch.SubmitJobInput) (*batch.SubmitJobOutput, error) {
			capturedInput = input
			return &batch.SubmitJobOutput{
				JobId:   aws.String("test-job-id"),
				JobName: input.JobName,
			}, nil
		},
	}

	b := &Backend{
		client: mockClient,
		event:  &noopEventWriter{},
		log:    log,
		conf: &config.AWSBatch{
			JobDefinition: "funnel-job-def",
			JobQueue:      "funnel-job-queue",
		},
	}

	task := &tes.Task{
		Id: "task1",
		Executors: []*tes.Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	err := b.Submit(task)
	if err != nil {
		t.Fatal(err)
	}

	if capturedInput == nil {
		t.Fatal("expected SubmitJob to be called")
	}
	if *capturedInput.JobDefinition != "funnel-job-def" {
		t.Errorf("expected job definition %q, got %q", "funnel-job-def", *capturedInput.JobDefinition)
	}
	if *capturedInput.JobQueue != "funnel-job-queue" {
		t.Errorf("expected job queue %q, got %q", "funnel-job-queue", *capturedInput.JobQueue)
	}
	if capturedInput.Parameters["taskID"] == nil || *capturedInput.Parameters["taskID"] != "task1" {
		t.Errorf("expected taskID parameter %q, got %v", "task1", capturedInput.Parameters["taskID"])
	}
}
