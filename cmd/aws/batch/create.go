package batch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

func init() {
	c := createCmd.Flags()
	c.StringVar(&conf.JobDef.Image, "container", conf.JobDef.Image,
		"Funnel worker Docker container to run")
	c.StringVar(&conf.Region, "region", conf.Region,
		"Region in which to create the Batch resources")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a compute environment, job queue, and job definition in a specified region",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New("batch create cmd")

		cli, err := newBatchSvc(conf, false)
		if err != nil {
			return err
		}
		batchCli := batch.New(cli.sess)

		a, err := batchCli.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{aws.String(conf.ComputeEnv.Name)},
		})
		if err != nil {
			return err
		}
		if len(a.ComputeEnvironments) == 0 {
			r, err := cli.CreateComputeEnvironment()
			if err != nil {
				return err
			}
			log.Info("Created ComputeEnvironment",
				"Name", *r.ComputeEnvironmentName,
				"Arn", *r.ComputeEnvironmentArn,
			)
		} else {
			log.Error("ComputeEnvironment already exists",
				"Name", *a.ComputeEnvironments[0].ComputeEnvironmentName,
				"Arn", *a.ComputeEnvironments[0].ComputeEnvironmentArn,
			)
		}

		b, err := batchCli.DescribeJobQueues(&batch.DescribeJobQueuesInput{
			JobQueues: []*string{aws.String(conf.JobQueue.Name)},
		})
		if err != nil {
			return err
		}
		if len(b.JobQueues) == 0 {
			r, err := cli.CreateJobQueue()
			if err != nil {
				return err
			}
			log.Info("Created JobQueue",
				"Name", *r.JobQueueName,
				"Arn", *r.JobQueueArn,
			)
		} else {
			log.Error("JobQueue already exists",
				"Name", *b.JobQueues[0].JobQueueName,
				"Arn", *b.JobQueues[0].JobQueueArn,
			)
		}

		c, err := batchCli.DescribeJobDefinitions(&batch.DescribeJobDefinitionsInput{
			JobDefinitionName: aws.String(conf.JobDef.Name),
			Status:            aws.String("ACTIVE"),
		})
		if err != nil {
			return err
		}
		if len(c.JobDefinitions) == 0 {
			r, err := cli.CreateJobDef()
			if err != nil {
				return err
			}
			log.Info("Registered JobDefinition",
				"Name", *r.JobDefinitionName,
				"Arn", *r.JobDefinitionArn,
			)
		} else {
			log.Error("JobDefinition already exists",
				"Name", *c.JobDefinitions[0].JobDefinitionName,
				"Arn", *c.JobDefinitions[0].JobDefinitionArn,
			)
		}

		return nil
	},
}
