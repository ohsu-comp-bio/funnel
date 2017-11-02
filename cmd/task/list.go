package task

import (
	"encoding/json"
	"fmt"
	"github.com/elgs/jsonql"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"io"
	"strings"
)

// List runs the "task list" CLI command, which connects to the server,
// calls ListTasks() and requests the given task view. Results may be filtered
// client-side using the "taskState" and "taskName" args. Output is written
// to the given writer.
func List(server, taskListView, taskState, taskName string, writer io.Writer) error {
	cli := client.NewClient(server)

	view, ok := tes.TaskView_value[taskListView]
	if !ok {
		return fmt.Errorf("Unknown task view: %s", taskListView)
	}

	var page string
	var taskArr []interface{}
	for {
		resp, err := cli.ListTasks(context.Background(), &tes.ListTasksRequest{
			View:      tes.TaskView(view),
			PageToken: page,
		})
		if err != nil {
			return err
		}
		if len(resp.Tasks) == 0 {
			break
		}
		page = resp.NextPageToken
		// convert resp to map[string]interface{} for query
		var out map[string]interface{}
		j, _ := cli.Marshaler.MarshalToString(resp)
		_ = json.Unmarshal([]byte(j), &out)
		taskArr = append(taskArr, out["tasks"].([]interface{})...)
		if page == "" {
			break
		}
	}

	// query tasks
	var tasks interface{}
	tasks = taskArr
	parser := jsonql.NewQuery(tasks)
	var queries []string

	if taskState != "" {
		queries = append(queries, fmt.Sprintf("state~='%s'", taskState))
	}
	if taskName != "" {
		queries = append(queries, fmt.Sprintf("name~='%s'", taskName))
	}
	if len(queries) > 0 {
		var err error
		tasks, err = parser.Query(strings.Join(queries, " && "))
		if err != nil {
			return err
		}
	}

	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	response := string(tasksJSON)
	if response == "null" {
		response = "{}"
	}

	fmt.Fprintf(writer, "%s\n", response)
	return nil
}
