package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <task_id> ...",
	Short: "get one or more tasks by ID",
	Run: func(cmd *cobra.Command, args []string) {
		for _, taskID := range args {
			resp, err := http.Get(tesServer + "/v1/jobs/" + taskID)
			body := responseChecker(resp, err)
			fmt.Printf("%s\n", body)
		}
	},
}

func init() {
	taskCmd.AddCommand(getCmd)
}
