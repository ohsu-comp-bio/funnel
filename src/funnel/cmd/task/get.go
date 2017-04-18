package task

import (
	"fmt"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [taskID ...]",
	Short: "get one or more tasks by ID",
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
	TaskCmd.AddCommand(getCmd)
}

func doGet(server string, ids []string) ([]string, error) {
	client := NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		body, err := client.GetTask(taskID)
		if err != nil {
			return nil, err
		}
		res = append(res, string(body))
	}
	return res, nil
}
