package batch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"regexp"
	"time"
)

var log = logger.Sub("batch")

// NewBackend returns a new local Backend instance.
func NewBackend(conf config.AWSBatch) (*Backend, error) {
	sess, err := util.NewAWSSession(conf.Key, conf.Secret, conf.Region)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating batch client: %v", err)
	}

	batchCli := batch.New(sess)

	response, err := batchCli.RegisterJobDefinition(&batch.RegisterJobDefinitionInput{
		ContainerProperties: &batch.ContainerProperties{
			Image:      aws.String(conf.JobDef.Image),
			Memory:     aws.Int64(conf.JobDef.DefaultMemory),
			Vcpus:      aws.Int64(conf.JobDef.DefaultVcpus),
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
				aws.String("worker"),
				aws.String("run"),
				aws.String("--task-id"),
				// This is a template variable that will be replaced with the taskID.
				aws.String("Ref::taskID"),
			},
		},
		JobDefinitionName: aws.String(conf.JobDef.Name),
		Type:              aws.String("container"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create base funnel job definition: %v", err)
	}

	return &Backend{
		client: batchCli,
		conf:   conf,
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
	log.Debug("Submitting to batch", "taskID", task.Id)

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
