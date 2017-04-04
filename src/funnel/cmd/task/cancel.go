package task

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"os"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel [taskID ...]",
	Short: "cancel one or more tasks by ID",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
		for _, taskID := range args {
			u, err := url.Parse(tesServer + "/v1/jobs/" + taskID)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			cli := &http.Client{}
			resp, err := cli.Do(&http.Request{
				Method: "DELETE",
				URL:    u,
			})
			body := responseChecker(resp, err)
			fmt.Printf("%s\n", body)
		}
	},
}

func init() {
	TaskCmd.AddCommand(cancelCmd)
}
