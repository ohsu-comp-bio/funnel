package task

import (
	"fmt"
	"github.com/spf13/cobra"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel [taskID ...]",
	Short: "cancel one or more tasks by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		res, err := doCancel(tesServer, args)
		if err != nil {
			return err
		}

		for _, x := range res {
			fmt.Println(x)
		}
		return nil
	},
}

func doCancel(server string, ids []string) ([]string, error) {
	client := NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		resp, err := client.CancelTask(taskID)
		if err != nil {
			return nil, err
		}
		// CancelTaskResponse is an empty struct
		out, err := client.marshaler.MarshalToString(resp)
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}
