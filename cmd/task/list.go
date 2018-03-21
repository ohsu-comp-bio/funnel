package task

import (
	"fmt"
	"io"
	"strings"

	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// List runs the "task list" CLI command, which connects to the server,
// calls ListTasks() and requests the given task view.
// Output is written to the given writer.
func List(server, taskView, pageToken, stateFilter string, tagsFilter []string, pageSize uint32, all bool, writer io.Writer) error {
	cli, err := tes.NewClient(server)
	if err != nil {
		return err
	}

	view, err := getTaskView(taskView)
	if err != nil {
		return err
	}

	output := &tes.ListTasksResponse{}

	state, err := getTaskState(stateFilter)
	if err != nil {
		return err
	}

	tags := make(map[string]string)
	for _, v := range tagsFilter {
		parts := strings.Split(v, "=")
		if len(parts) != 2 {
			return fmt.Errorf("tags must be of the form: KEY=VALUE")
		}
		tags[parts[0]] = parts[1]
	}

	for {
		req := &tes.ListTasksRequest{
			View:      tes.TaskView(view),
			PageToken: pageToken,
			PageSize:  pageSize,
			State:     state,
			Tags:      tags,
		}

		resp, err := cli.ListTasks(context.Background(), req)
		if err != nil {
			return err
		}

		output.Tasks = append(output.Tasks, resp.Tasks...)
		output.NextPageToken = resp.NextPageToken
		pageToken = resp.NextPageToken

		if !all || (all && pageToken == "") {
			break
		}
	}

	response, err := cli.Marshaler.MarshalToString(output)
	if err != nil {
		return fmt.Errorf("marshaling error: %v", err)
	}

	fmt.Fprintf(writer, "%s\n", response)
	return nil
}
