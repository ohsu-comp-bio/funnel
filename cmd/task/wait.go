package task

import (
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/spf13/cobra"
)

var waitCmd = &cobra.Command{
	Use:   "wait [taskID...]",
	Short: "Wait for one or more tasks to complete.\n",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPerID(cmd, args, doWait)
	},
}

func doWait(cli *client.Client, id string) error {
	err := cli.WaitForTask(id)
	if err != nil {
		return err
	}
	return nil
}
