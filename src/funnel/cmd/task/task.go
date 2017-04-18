package task

import (
	"funnel/logger"
	"github.com/spf13/cobra"
)

var tesServer string
var log = logger.New("task-cmd")

// TaskCmd represents the task command
var TaskCmd = &cobra.Command{
	Use: "task",
}

func init() {
	TaskCmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
}
