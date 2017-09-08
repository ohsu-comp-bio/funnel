package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/spf13/cobra"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel [taskID ...]",
	Short: "cancel one or more tasks by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPerID(cmd, args, doCancel)
	},
}

func doCancel(cli *client.Client, id string) error {
	resp, err := cli.CancelTask(id)
	if err != nil {
		return err
	}
	// CancelTaskResponse is an empty struct
	x, err := cli.Marshaler.MarshalToString(resp)
	if err != nil {
		return err
	}

	fmt.Println(x)
	return nil
}
