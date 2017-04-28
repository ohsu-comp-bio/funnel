package task

import (
	"encoding/json"
	"fmt"
	"github.com/elgs/jsonql"
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := doList(tesServer)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", r)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&taskState, "state", "s", "", "Task state")
	listCmd.Flags().StringVarP(&taskName, "name", "n", "", "Task name")
}

func doList(server string) (string, error) {
	client := NewClient(server)
	resp, err := client.ListTasks()
	if err != nil {
		return "", err
	}

	// convert resp to map[string]interface{} for query
	var out map[string]interface{}
	j, _ := client.marshaler.MarshalToString(resp)
	_ = json.Unmarshal([]byte(j), &out)
	tasks := out["tasks"]

	// query tasks
	parser := jsonql.NewQuery(tasks)
	var queries []string

	if taskState != "" {
		queries = append(queries, fmt.Sprintf("state~='%s'", taskState))
	}
	if taskName != "" {
		queries = append(queries, fmt.Sprintf("name~='%s'", taskName))
	}
	if len(queries) > 0 {
		tasks, err = parser.Query(strings.Join(queries, " && "))
		if err != nil {
			return "", err
		}
	}

	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		return "", err
	}
	return string(tasksJSON), nil
}
