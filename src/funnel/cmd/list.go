package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

var (
	taskName  string
	taskDesc  string
	taskTags  string
	taskStart string
	taskEnd   string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all tasks",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(tesServer + "/v1/jobs")
		body := responseChecker(resp, err)
		fmt.Printf("%s", body)
	},
}

func init() {
	taskCmd.AddCommand(listCmd)
}
