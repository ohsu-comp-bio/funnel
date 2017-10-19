package batch

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

func init() {
	f := updateCmd.Flags()
	f.StringVar(&funnelConfigFile, "config", funnelConfigFile, "Funnel configuration file")
	f.StringVar(&conf.Region, "region", conf.Region, "Region in which to create the Batch resources")
	f.StringVar(&conf.JobDef.Name, "JobDef.Name", conf.JobDef.Name, "The name of the job definition.")
	f.StringVar(&conf.JobDef.Image, "JobDef.Image", conf.JobDef.Image, "The docker image used to start a container.")
	f.Int64Var(&conf.JobDef.MemoryMiB, "JobDef.MemoryMiB", conf.JobDef.MemoryMiB, "The hard limit (in MiB) of memory to present to the container.")
	f.Int64Var(&conf.JobDef.VCPUs, "JobDef.VCPUs", conf.JobDef.VCPUs, "The number of vCPUs reserved for the container.")
	f.StringVar(&conf.JobDef.JobRoleArn, "JobDef.JobRoleArn", conf.JobDef.JobRoleArn, "The Amazon Resource Name (ARN) of the IAM role that the container can assume for AWS permissions. A role will be created if not provided.")
}

var updateCmd = &cobra.Command{
	Use:   "update-job-def",
	Short: "Revise a job definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.NewLogger("batch-revise-job-def", logger.DefaultConfig())

		if funnelConfigFile != "" {
			funnelConf := config.Config{}
			config.ParseFile(funnelConfigFile, &funnelConf)
			conf.FunnelWorker = funnelConf.Worker
		}

		cli, err := newBatchSvc(conf)
		if err != nil {
			return err
		}

		c, err := cli.CreateJobDefinition(true)
		if err != nil {
			return fmt.Errorf("failed to revise JobDefinition: %v", err)
		}
		log.Info("Created JobDefinition",
			"Name", *c.JobDefinitionName,
			"Arn", *c.JobDefinitionArn,
			"Revision", *c.Revision,
		)

		return nil
	},
}
