package cmd

import (
	"funnel/cmd/gce"
	"funnel/cmd/server"
	"funnel/cmd/task"
	"funnel/cmd/worker"
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
}
