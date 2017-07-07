package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func newBatchClient(conf Config) *batchsvc {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(conf.Region),
	}))

	return &batchsvc{
		conf:           conf,
		batch:          batch.New(sess),
		cloudwatchlogs: cloudwatchlogs.New(sess),
	}
}

type batchsvc struct {
	conf           Config
	batch          Batch
	cloudwatchlogs CloudWatchLogs
}

func (b *batchsvc) CreateJob(task *tes.Task) (*batch.SubmitJobOutput, error) {

	marshaler := jsonpb.Marshaler{}
	taskJSON, err := marshaler.MarshalToString(task)
	if err != nil {
		return nil, err
	}

	return b.batch.SubmitJob(&batch.SubmitJobInput{
		JobDefinition: aws.String(b.conf.JobDef.Name),
		JobName:       aws.String(safeJobName(task.Name)),
		JobQueue:      aws.String(b.conf.JobQueue.Name),
		Parameters: map[string]*string{
			// Include the entire task message, encoded as a JSON string,
			// in the job parameters. This gets used by the AWS Batch
			// task runner.
			"task": aws.String(taskJSON),
		},
	})
}

func (b *batchsvc) DescribeJob(id string) (*batch.DescribeJobsOutput, error) {
	return b.batch.DescribeJobs(&batch.DescribeJobsInput{
		Jobs: []*string{
			aws.String(id),
		},
	})
}

func (b *batchsvc) TerminateJob(id string) (*batch.TerminateJobOutput, error) {
	return b.batch.TerminateJob(&batch.TerminateJobInput{
		JobId:  aws.String(id),
		Reason: aws.String(cancelReason),
	})
}

func (b *batchsvc) ListJobs(status, token string, size int64) (*batch.ListJobsOutput, error) {
	return b.batch.ListJobs(&batch.ListJobsInput{
		JobQueue:   aws.String(b.conf.JobQueue.Name),
		JobStatus:  aws.String(status),
		MaxResults: aws.Int64(size),
		NextToken:  aws.String(token),
	})
}

func (b *batchsvc) CreateComputeEnvironment() (*batch.CreateComputeEnvironmentOutput, error) {

	conf := b.conf.ComputeEnv
	return b.batch.CreateComputeEnvironment(&batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(conf.Name),
		ComputeResources: &batch.ComputeResource{
			InstanceRole:     aws.String(conf.InstanceRole),
			InstanceTypes:    convertStringSlice(conf.InstanceTypes),
			MaxvCpus:         aws.Int64(conf.MaxVCPUs),
			MinvCpus:         aws.Int64(conf.MinVCPUs),
			SecurityGroupIds: convertStringSlice(conf.SecurityGroupIds),
			Subnets:          convertStringSlice(conf.Subnets),
			Tags:             convertStringMap(conf.Tags),
			Type:             aws.String("EC2"),
		},
		ServiceRole: aws.String(conf.ServiceRole),
		State:       aws.String("ENABLED"),
		Type:        aws.String("MANAGED"),
	})
}

func (b *batchsvc) CreateJobQueue() (*batch.CreateJobQueueOutput, error) {
	conf := b.conf.JobQueue

	var envs []*batch.ComputeEnvironmentOrder
	for i, c := range conf.ComputeEnvs {
		envs = append(envs, &batch.ComputeEnvironmentOrder{
			ComputeEnvironment: aws.String(c),
			Order:              aws.Int64(int64(i)),
		})
	}

	return b.batch.CreateJobQueue(&batch.CreateJobQueueInput{
		ComputeEnvironmentOrder: envs,
		JobQueueName:            aws.String(conf.Name),
		Priority:                aws.Int64(1),
		State:                   aws.String("ENABLED"),
	})
}

func (b *batchsvc) CreateJobDef() (*batch.RegisterJobDefinitionOutput, error) {
	conf := b.conf.JobDef

	return b.batch.RegisterJobDefinition(&batch.RegisterJobDefinitionInput{
		ContainerProperties: &batch.ContainerProperties{
			Image:      aws.String(b.conf.Container),
			Memory:     aws.Int64(conf.Memory),
			Vcpus:      aws.Int64(conf.VCPUs),
			Privileged: aws.Bool(true),
			MountPoints: []*batch.MountPoint{
				{
					SourceVolume:  aws.String("docker_sock"),
					ContainerPath: aws.String("/var/run/docker.sock"),
				},
			},
			Volumes: []*batch.Volume{
				{
					Name: aws.String("docker_sock"),
					Host: &batch.Host{
						SourcePath: aws.String("/var/run/docker.sock"),
					},
				},
			},
			Command: []*string{
				aws.String("aws"),
				aws.String("runtask"),
				aws.String("--task"),
				// This is a template variable that will be replaced with
				// the full TES task message in JSON form.
				aws.String("Ref::task"),
			},
		},
		JobDefinitionName: aws.String(b.conf.JobDef.Name),
		Type:              aws.String("container"),
	})
}

func (b *batchsvc) GetTaskLogs(name, jobID, attemptID string) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return b.cloudwatchlogs.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("/aws/batch/job"),
		LogStreamName: aws.String(name + "/" + jobID + "/" + attemptID),
		StartFromHead: aws.Bool(true),
	})
}

func logErr(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case batch.ErrCodeClientException:
			log.Error(batch.ErrCodeClientException, aerr.Error())
		case batch.ErrCodeServerException:
			log.Error(batch.ErrCodeServerException, aerr.Error())
		default:
			log.Error("Error", aerr.Error())
		}
	} else {
		log.Error("Error", err)
	}
}

func convertStringSlice(s []string) []*string {
	var ret []*string
	for _, t := range s {
		ret = append(ret, aws.String(t))
	}
	return ret
}

func convertStringMap(s map[string]string) map[string]*string {
	m := map[string]*string{}
	for k, v := range s {
		m[k] = aws.String(v)
	}
	return m
}
