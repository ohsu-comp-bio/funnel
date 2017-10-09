package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
)

func Get(server string, ids []string, taskView string) error {
	cli := client.NewClient(server)
	res := []string{}

	for _, taskID := range ids {
		resp, err := cli.GetTask(taskID, taskView)
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
		fmt.Println(x)
	}
	return nil
}
