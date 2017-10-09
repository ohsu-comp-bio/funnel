package task

import (
	"encoding/json"
	"fmt"
	"github.com/elgs/jsonql"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strings"
)

func List(server, taskListView, taskState, taskName string) error {
	cli := client.NewClient(server)

	view, ok := tes.TaskView_value[taskListView]
	if !ok {
		return fmt.Errorf("Unknown task view: %s", taskListView)
	}

	var page string
	var taskArr []interface{}
	for {
		resp, err := cli.ListTasks(&tes.ListTasksRequest{
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

	fmt.Printf("%s\n", response)
	return nil
}
