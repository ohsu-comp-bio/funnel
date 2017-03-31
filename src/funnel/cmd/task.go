package cmd

import (
	"github.com/spf13/cobra"
)

var tesServer string

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use: "task",
}

func init() {
	RootCmd.AddCommand(taskCmd)
	taskCmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
}
