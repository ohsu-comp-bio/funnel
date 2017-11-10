package batch

import (
	awsutil "github.com/ohsu-comp-bio/funnel/cmd/aws/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"time"
)

// Config represents configuration of the AWS proxy, including
// the compute environment, job queue, and base job definition.
type Config struct {
	Region       string
	ComputeEnv   ComputeEnvConfig
	JobQueue     JobQueueConfig
	JobDef       JobDefinitionConfig
	JobRole      JobRoleConfig
	FunnelWorker config.Worker
}

// ComputeEnvConfig represents configuration of the AWS Batch
// Compute Environment.
type ComputeEnvConfig struct {
	Name             string
	MinVCPUs         int64
	MaxVCPUs         int64
	SecurityGroupIds []string
	Subnets          []string
	Tags             map[string]string
	ServiceRole      string
	InstanceRole     string
	InstanceTypes    []string
}

// JobQueueConfig represents configuration of the AWS Batch
// Job Queue.
type JobQueueConfig struct {
	Name        string
	Priority    int64
	ComputeEnvs []string
}

// JobDefinitionConfig represents configuration of the AWS Batch Job Definition
type JobDefinitionConfig struct {
	Name       string
	Image      string
	MemoryMiB  int64
	VCPUs      int64
	JobRoleArn string
}

// JobRoleConfig represents configuration of the AWS Batch
// JobRole.
type JobRoleConfig struct {
	RoleName           string
	S3PolicyName       string
	DynamoDBPolicyName string
	Policies           struct {
		AssumeRole awsutil.AssumeRolePolicy
		S3         awsutil.Policy
		DynamoDB   awsutil.Policy
	}
}

// DefaultConfig returns default configuration of for AWS Batch resource creation.
func DefaultConfig() Config {
	c := Config{
		Region: "",
		ComputeEnv: ComputeEnvConfig{
			Name:          "funnel-compute-environment",
			InstanceTypes: []string{"optimal"},
			MinVCPUs:      0,
			MaxVCPUs:      256,
			Tags: map[string]string{
				"Name": "Funnel",
			},
		},
		JobQueue: JobQueueConfig{
			Name:     "funnel-job-queue",
			Priority: 1,
			ComputeEnvs: []string{
				"funnel-compute-environment",
			},
		},
		JobRole: JobRoleConfig{
			RoleName:           "FunnelECSTaskRole",
			DynamoDBPolicyName: "FunnelDynamoDB",
			S3PolicyName:       "FunnelS3",
		},
		JobDef: JobDefinitionConfig{
			Name:      "funnel-job-def",
			Image:     "docker.io/ohsucompbio/funnel:latest",
			VCPUs:     1,
			MemoryMiB: 128,
		},
	}

	c.JobRole.Policies.AssumeRole = awsutil.AssumeRolePolicy{
		Version: "2012-10-17",
		Statement: []awsutil.RoleStatement{
			{
				Effect:    "Allow",
				Principal: map[string]string{"Service": "ecs-tasks.amazonaws.com"},
				Action:    "sts:AssumeRole",
			},
		},
	}
	c.JobRole.Policies.S3 = awsutil.Policy{
		Version: "2012-10-17",
		Statement: []awsutil.Statement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:GetObject",
					"s3:ListObjects",
					"s3:ListBucket",
					"s3:CreateBucket",
					"s3:PutObject",
				},
				Resource: "*",
			},
		},
	}
	c.JobRole.Policies.DynamoDB = awsutil.Policy{
		Version: "2012-10-17",
		Statement: []awsutil.Statement{
			{
				Effect: "Allow",
				Action: []string{
					"dynamodb:GetItem",
					"dynamodb:PutItem",
					"dynamodb:UpdateItem",
					"dynamodb:Query",
				},
				Resource: "*",
			},
		},
	}

	c.FunnelWorker.WorkDir = "/opt/funnel-work-dir"
	c.FunnelWorker.UpdateRate = time.Second * 10
	c.FunnelWorker.BufferSize = 10000
	c.FunnelWorker.TaskReader = "dynamodb"
	c.FunnelWorker.TaskReaders.DynamoDB.TableBasename = "funnel"
	c.FunnelWorker.TaskReaders.DynamoDB.Region = ""
	c.FunnelWorker.ActiveEventWriters = []string{"dynamodb", "log"}
	c.FunnelWorker.EventWriters.DynamoDB.TableBasename = "funnel"
	c.FunnelWorker.EventWriters.DynamoDB.Region = ""

	return c
}
