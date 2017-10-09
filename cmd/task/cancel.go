package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
)

func Cancel(server string, ids []string) error {
	cli := client.NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		resp, err := cli.CancelTask(taskID)
		if err != nil {
			return err
		}
		// CancelTaskResponse is an empty struct
		out, err := cli.Marshaler.MarshalToString(resp)
		if err != nil {
			return err
		}
		res = append(res, out)
	}

	for _, x := range res {
		fmt.Println(x)
	}
	return nil
}
