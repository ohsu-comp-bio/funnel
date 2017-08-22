package task

import (
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/spf13/cobra"
)

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait [taskID...]",
	Short: "Wait for one or more tasks to complete.",
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
