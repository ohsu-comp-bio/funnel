package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/elgs/jsonql"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strings"
)

var (
	taskName  string
	taskState string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all tasks",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(tesServer + "/v1/jobs")
		body := responseChecker(resp, err)

		var jobList map[string]interface{}
		err = json.Unmarshal(body, &jobList)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		jobs := jobList["jobs"]
		parser := jsonql.NewQuery(jobs)
		var queries []string

		if taskState != "" {
			queries = append(queries, fmt.Sprintf("state~='%s'", taskState))
		}
		if taskName != "" {
			queries = append(queries, fmt.Sprintf("task.name~='%s'", taskName))
		}
		if len(queries) > 0 {
			jobs, err = parser.Query(strings.Join(queries, " && "))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		jobsJson, err := json.Marshal(jobs)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("%s", jobsJson)
	},
}

func init() {
	taskCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&taskState, "state", "s", "", "Task state")
	listCmd.Flags().StringVarP(&taskName, "name", "n", "", "Task name")
}
