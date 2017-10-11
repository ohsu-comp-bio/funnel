package batch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"regexp"
	"time"
)

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.Config) (*Backend, error) {
	batchConf := conf.Backends.Batch

	awsConf := util.NewAWSConfigWithCreds(batchConf.Credentials.Key, batchConf.Credentials.Secret)
	awsConf.WithRegion(batchConf.Region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating batch client: %v", err)
	}

	batchCli := batch.New(sess)

	jobDef := &batch.RegisterJobDefinitionInput{
		ContainerProperties: &batch.ContainerProperties{
			Image:      aws.String(batchConf.JobDef.Image),
			Memory:     aws.Int64(batchConf.JobDef.DefaultMemory),
			Vcpus:      aws.Int64(batchConf.JobDef.DefaultVcpus),
			Privileged: aws.Bool(true),
			MountPoints: []*batch.MountPoint{
				{
					SourceVolume:  aws.String("docker_sock"),
					ContainerPath: aws.String("/var/run/docker.sock"),
				},
				{
					SourceVolume:  aws.String("funnel-work-dir"),
					ContainerPath: aws.String(conf.Worker.WorkDir),
				},
			},
			Volumes: []*batch.Volume{
				{
					Name: aws.String("docker_sock"),
					Host: &batch.Host{
						SourcePath: aws.String("/var/run/docker.sock"),
					},
				},
				{
					Name: aws.String("funnel-work-dir"),
					Host: &batch.Host{
						SourcePath: aws.String(conf.Worker.WorkDir),
					},
				},
			},
			Command: []*string{
				aws.String("worker"),
				aws.String("run"),
				aws.String("--WorkDir"),
				aws.String(conf.Worker.WorkDir),
				aws.String("--TaskReader"),
				aws.String(conf.Worker.TaskReader),
				aws.String("--DynamoDB.Region"),
				aws.String(conf.Worker.EventWriters.DynamoDB.Region),
				aws.String("--DynamoDB.TableBasename"),
				aws.String(conf.Worker.EventWriters.DynamoDB.TableBasename),
				aws.String("--task-id"),
				// This is a template variable that will be replaced with the taskID.
				aws.String("Ref::taskID"),
			},
			JobRoleArn: aws.String(batchConf.JobDef.JobRoleArn),
		},
		JobDefinitionName: aws.String(batchConf.JobDef.Name),
		Type:              aws.String("container"),
	}
	for _, val := range conf.Worker.ActiveEventWriters {
		jobDef.ContainerProperties.Command = append(jobDef.ContainerProperties.Command, aws.String("--ActiveEventWriters"), aws.String(val))
	}

	response, err := batchCli.RegisterJobDefinition(jobDef)
	if err != nil {
		return nil, fmt.Errorf("failed to create base funnel job definition: %v", err)
	}

	return &Backend{
		client: batchCli,
		conf:   conf.Backends.Batch,
		jobDef: *response.JobDefinitionArn,
	}, nil
}

// Backend represents the local backend.
type Backend struct {
	client *batch.Batch
	conf   config.AWSBatch
	jobDef string
}

// Submit submits a task to the AWS batch service.
func (b *Backend) Submit(task *tes.Task) error {

	req := &batch.SubmitJobInput{
		JobDefinition: aws.String(b.jobDef),
		JobName:       aws.String(safeJobName(task.Name)),
		JobQueue:      aws.String(b.conf.JobQueue),
		Parameters: map[string]*string{
			// Include the taskID in the job parameters. This gets used by
			// the funnel 'worker run' cmd.
			"taskID": aws.String(task.Id),
		},
	}

	// convert ram from GB to MiB
	ram := int64(task.Resources.RamGb * 953.674)
	vcpus := int64(task.Resources.CpuCores)
	if ram > 0 {
		req.ContainerOverrides.Memory = aws.Int64(ram)
	}

	if vcpus > 0 {
		req.ContainerOverrides.Vcpus = aws.Int64(vcpus)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := b.client.SubmitJobWithContext(ctx, req)
	return err
}

// AWS limits the characters allowed in job names,
// so replace invalid characters with underscores.
func safeJobName(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	return re.ReplaceAllString(s, "_")
}
