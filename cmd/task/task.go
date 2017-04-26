package task

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

var tesServer string
var log = logger.New("task cmd")

// Cmd represents the task command
var Cmd = &cobra.Command{
	Use: "task",
}

func init() {
	Cmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(cancelCmd)
}
