package task

import (
	"fmt"
	"io"

	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// Cancel runs the "task cancel" CLI command, which connects to the server,
// calls CancelTask() on each ID, and writes output to the given writer.
func Cancel(server string, ids []string, writer io.Writer) error {
	cli, err := tes.NewClient(server)
	if err != nil {
		return err
	}

	res := []string{}

	for _, taskID := range ids {
		resp, err := cli.CancelTask(context.Background(), &tes.CancelTaskRequest{Id: taskID})
		if err != nil {
			return err
		}
		// CancelTaskResponse is an empty struct
		out := cli.Marshaler.Format(resp)
		res = append(res, out)
	}

	for _, x := range res {
		fmt.Fprintln(writer, x)
	}
	return nil
}
