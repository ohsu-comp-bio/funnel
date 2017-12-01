package batch

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
	"regexp"
	"time"
)

// NewBackend returns a new local Backend instance.
func NewBackend(batchConf config.AWSBatch) (*Backend, error) {

	sess, err := util.NewAWSSession(batchConf.AWSConfig)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating batch client: %v", err)
	}

	return &Backend{
		client: batch.New(sess),
		conf:   batchConf,
	}, nil
}

// Backend represents the local backend.
type Backend struct {
	client *batch.Batch
	conf   config.AWSBatch
}

// Submit submits a task to the AWS batch service.
func (b *Backend) Submit(task *tes.Task) error {
	req := &batch.SubmitJobInput{
		// JobDefinition: aws.String(b.jobDef),
		JobDefinition: aws.String(b.conf.JobDefinition),
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
