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
	Short: "Get one or more tasks by ID.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		res, err := doGet(tesServer, args)
		if err != nil {
			return err
		}

		for _, x := range res {
			fmt.Println(x)
		}

		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&taskView, "view", "v", "FULL", "Task view")
}

func doGet(server string, ids []string) ([]string, error) {
	cli := client.NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		resp, err := cli.GetTask(taskID, taskView)
		if err != nil {
			return nil, err
		}
		out, err := cli.Marshaler.MarshalToString(resp)
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}
