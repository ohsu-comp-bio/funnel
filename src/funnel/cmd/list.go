package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
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
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%s", body)
	},
}

func init() {
	taskCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&taskName, "name", "n", "", "Task name")
	listCmd.Flags().StringVarP(&taskDesc, "desc", "d", "", "Task description")
	listCmd.Flags().StringVarP(&taskTags, "tags", "t", "", "Task tags")
	listCmd.Flags().StringVarP(&taskStart, "start", "s", "", "Task start time")
	listCmd.Flags().StringVarP(&taskEnd, "end", "e", "", "Task end time")
}
