package batch

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

func init() {
	c := createCmd.Flags()
	c.StringVar(&conf.Region, "region", conf.Region,
		"Region in which to create the Batch resources")
	c.StringVar(&conf.ComputeEnv.Name, "ComputeEnv.Name", conf.ComputeEnv.Name,
		"The name of the compute environment.")
	c.Int64Var(&conf.ComputeEnv.MinVCPUs, "ComputEnv.MinVCPUs", conf.ComputeEnv.MinVCPUs,
		"The minimum number of EC2 vCPUs that an environment should maintain. (default 0)")
	c.Int64Var(&conf.ComputeEnv.MaxVCPUs, "ComputEnv.MaxVCPUs", conf.ComputeEnv.MaxVCPUs,
		"The maximum number of EC2 vCPUs that an environment can reach.")
	c.StringSliceVar(&conf.ComputeEnv.SecurityGroupIds, "ComputEnv.SecurityGroupIds", conf.ComputeEnv.SecurityGroupIds,
		"The EC2 security groups that are associated with instances launched in the compute environment. If none are specified all security groups will be used.")
	c.StringSliceVar(&conf.ComputeEnv.Subnets, "ComputEnv.Subnets", conf.ComputeEnv.Subnets,
		"The VPC subnets into which the compute resources are launched. If none are specified all subnets will be used.")
	c.StringSliceVar(&conf.ComputeEnv.InstanceTypes, "ComputEnv.InstanceTypes", conf.ComputeEnv.InstanceTypes,
		"The instances types that may be launched. You can also choose optimal to pick instance types on the fly that match the demand of your job queues.")
	c.StringVar(&conf.JobQueue.Name, "JobQueue.Name", conf.JobQueue.Name,
		"The name of the job queue.")
	c.Int64Var(&conf.JobQueue.Priority, "JobQueue.Priority", conf.JobQueue.Priority,
		"The priority of the job queue. Priority is determined in descending order.")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a compute environment, job queue and job role in a specified region",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createBatchResources(conf)
	},
}

func createBatchResources(conf Config) error {
	log := logger.NewLogger("batch-create-resources", logger.DefaultConfig())

	cli, err := newBatchSvc(conf)
	if err != nil {
		return err
	}

	a, err := cli.CreateComputeEnvironment()
	switch err.(type) {
	case nil:
		log.Info("Created ComputeEnvironment",
			"Name", *a.ComputeEnvironmentName,
			"Arn", *a.ComputeEnvironmentArn,
		)
	case *errResourceExists:
		log.Info("ComputeEnvironment already exists",
			"Name", *a.ComputeEnvironmentName,
			"Arn", *a.ComputeEnvironmentArn,
		)
	default:
		return fmt.Errorf("failed to create ComputeEnvironment: %v", err)
	}

	b, err := cli.CreateJobQueue()
	switch err.(type) {
	case nil:
		log.Info("Created JobQueue",
			"Name", *b.JobQueueName,
			"Arn", *b.JobQueueArn,
		)
	case *errResourceExists:
		log.Info("JobQueue already exists",
			"Name", *b.JobQueueName,
			"Arn", *b.JobQueueArn,
		)
	default:
		return fmt.Errorf("failed to create JobQueue: %v", err)
	}

	c, err := cli.CreateJobRole()
	switch err.(type) {
	case nil:
		log.Info("Created Role",
			"Name", *c.Role.RoleName,
			"Arn", *c.Role.Arn,
		)
	case *errResourceExists:
		log.Info("Role already exists",
			"Name", *c.Role.RoleName,
			"Arn", *c.Role.Arn,
		)
	default:
		return fmt.Errorf("failed to create JobRole: %v", err)
	}

	err = cli.AttachRolePolicies()
	switch err.(type) {
	case nil:
		log.Info("Attached inline policies to JobRole")
	case *errResourceExists:
		break
	default:
		return fmt.Errorf("failed to attach policies to JobRole: %v", err)
	}

	return nil
}
