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

func init() {
	TaskCmd.AddCommand(cancelCmd)
}

func doCancel(server string, ids []string) ([]string, error) {
	client := NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		body, err := client.CancelTask(taskID)
		if err != nil {
			return nil, err
		}
		res = append(res, string(body))
	}
	return res, nil
}
