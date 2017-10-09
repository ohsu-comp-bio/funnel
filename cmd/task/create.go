package task

import (
	"encoding/json"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"io/ioutil"
)

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func Create(server string, messages []string) error {
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
		fmt.Println(x)
	}

	return nil
}
