package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/spf13/cobra"
)

var taskView string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [taskID ...]",
	Short: "get one or more tasks by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPerID(cmd, args, doGet)
	},
}

func init() {
	getCmd.Flags().StringVarP(&taskView, "view", "v", "FULL", "Task view")
}

func doGet(cli *client.Client, id string) error {
	resp, err := cli.GetTask(id, taskView)
	if err != nil {
		return err
	}
	x, err := cli.Marshaler.MarshalToString(resp)
	if err != nil {
		return err
	}

	fmt.Println(x)
	return nil
}
