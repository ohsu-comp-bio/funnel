package task

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [task.json ...]",
	Short: "create one or more tasks to run on the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		res, err := doCreate(tesServer, args)
		if err != nil {
			return err
		}

		for _, x := range res {
			fmt.Println(x)
		}

		return nil
	},
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func doCreate(server string, messages []string) ([]string, error) {
	client := NewClient(server)
	res := []string{}

	for _, task := range messages {
		var err error
		var taskMessage []byte

		if isJSON(task) {
			taskMessage = []byte(task)
		} else {
			taskMessage, err = ioutil.ReadFile(task)
			if err != nil {
				return nil, err
			}
		}

		r, err := client.CreateTask(taskMessage)
		if err != nil {
			return nil, err
		}
		res = append(res, r.Id)
	}
	return res, nil
}
