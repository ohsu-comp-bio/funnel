package wait

import (
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/spf13/cobra"
)

var tesServer string

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "wait [taskID...]",
	Short: "Wait for one or more tasks to complete.\n",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		cli := client.NewClient(tesServer)

		err := cli.WaitForTask(args...)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
}
