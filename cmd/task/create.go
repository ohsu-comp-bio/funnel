package task

import (
	"encoding/json"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"io"
	"io/ioutil"
)

// Create runs the "task create" CLI command, connecting to the server,
// calling CreateTask, and writing output to the given writer.
// Tasks are loaded from the "messages" arg. "messages" may contain either
// file paths or JSON objects.
func Create(server string, messages []string, writer io.Writer) error {
	cli := client.NewClient(server)
	res := []string{}

	for _, task := range messages {
		var err error
		var taskMessage []byte

		if isJSON(task) {
			taskMessage = []byte(task)
		} else {
			taskMessage, err = ioutil.ReadFile(task)
			if err != nil {
				return err
			}
		}

		r, err := cli.CreateTask(taskMessage)
		if err != nil {
			return err
		}
		res = append(res, r.Id)
	}

	for _, x := range res {
		fmt.Fprintln(writer, x)
	}

	return nil
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
