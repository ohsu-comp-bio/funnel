package cmd

import (
	"funnel/cmd/task"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use: "funnel",
}

func init() {
	RootCmd.AddCommand(task.TaskCmd)
}
