package beanstalk

import (
	"github.com/spf13/cobra"
)

// Cmd represents the aws elastic beanstalk command
var Cmd = &cobra.Command{
	Use:   "ebs",
	Short: "Deploy funnel server on AWS Elastic Beanstalk",
}

func init() {
	Cmd.AddCommand(deployCmd)
}
