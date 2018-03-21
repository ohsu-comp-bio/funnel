package task

import (
	"fmt"
	"io"

	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// Get runs the "task get" CLI command, which connects to the server,
// calls GetTask for each ID, requesting the given task view, and writes
// output to the given writer.
func Get(server string, ids []string, taskView string, w io.Writer) error {
	cli, err := tes.NewClient(server)
	if err != nil {
		return err
	}

	res := []string{}

	view, err := getTaskView(taskView)
	if err != nil {
		return err
	}

	for _, taskID := range ids {
		resp, err := cli.GetTask(context.Background(), &tes.GetTaskRequest{
			Id:   taskID,
			View: tes.TaskView(view),
		})
		if err != nil {
			return err
		}
		out, err := cli.Marshaler.MarshalToString(resp)
		if err != nil {
			return err
		}
		res = append(res, out)
	}

	for _, x := range res {
		fmt.Fprintln(w, x)
	}
	return nil
}
