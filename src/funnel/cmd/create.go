package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"funnel/proto/tes"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <task.json>",
	Short: "create a task to run on the server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
		for _, task := range args {
			taskMessage, err := ioutil.ReadFile(task)
			if err != nil {
				if os.IsNotExist(err) {
					taskMessage = []byte(task)
				}
				fmt.Fprintf(os.Stderr, "File error: %v\n", err)
				os.Exit(1)
			}
			resp, err := http.Post(tesServer+"/v1/jobs", "application/json", bytes.NewReader(taskMessage))
			body := responseChecker(resp, err)
			var jobID = tes.JobID{}
			err = json.Unmarshal(body, &jobID)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("%s\n", jobID.Value)
		}
	},
}

func init() {
	taskCmd.AddCommand(createCmd)
}
