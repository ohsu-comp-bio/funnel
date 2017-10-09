package task

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
	"os"
)

var defaultTesServer = "http://localhost:8000"
var tesServer = defaultTesServer
var log = logger.New("task cmd")

// Cmd represents the task command
var Cmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Make API calls to a TES server.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if tesServer == defaultTesServer {
			tesServer = os.Getenv("FUNNEL_SERVER")
		}
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&tesServer, "server", "S", tesServer, "")
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(cancelCmd)
	Cmd.AddCommand(waitCmd)
}
