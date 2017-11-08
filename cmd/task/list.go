package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"io"
)

// List runs the "task list" CLI command, which connects to the server,
// calls ListTasks() and requests the given task view. Results may be filtered
// client-side using the "taskState" and "taskName" args. Output is written
// to the given writer.
func List(server, taskView, pageToken string, pageSize uint32, all bool, writer io.Writer) error {
	cli := client.NewClient(server)

	view, ok := tes.TaskView_value[taskView]
	if !ok {
		return fmt.Errorf("Unknown task view: %s", taskView)
	}

	output := &tes.ListTasksResponse{}

	for {
		resp, err := cli.ListTasks(context.Background(), &tes.ListTasksRequest{
			View:      tes.TaskView(view),
			PageToken: pageToken,
			PageSize:  pageSize,
		})
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
