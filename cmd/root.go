package cmd

import (
	"github.com/ohsu-comp-bio/funnel/cmd/examples"
	"github.com/ohsu-comp-bio/funnel/cmd/gce"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/cmd/task"
	"github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "funnel",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(worker.Cmd)
	RootCmd.AddCommand(server.Cmd)
	RootCmd.AddCommand(task.Cmd)
	RootCmd.AddCommand(gce.Cmd)
	RootCmd.AddCommand(examples.Cmd)
}
