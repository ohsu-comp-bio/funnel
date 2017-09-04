package task

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
	"os"
)

var tesServer string
var log = logger.New("task cmd")

// Cmd represents the task command
var Cmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Make API calls to a TES server.",
}

func init() {
	tesServer = os.Getenv("FUNNEL_SERVER")
	if tesServer == "" {
		tesServer = "http://localhost:8000"
	}

	Cmd.PersistentFlags().StringVarP(&tesServer, "server", "S", tesServer, `may be set by FUNNEL_SERVER env var`)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(cancelCmd)
	Cmd.AddCommand(waitCmd)
}
