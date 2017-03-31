package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get a task",
	Run: func(cmd *cobra.Command, args []string) {
		for _, jobID := range args {
			resp, err := http.Get(tesServer + "/v1/jobs/" + jobID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%s\n", body)
		}
	},
}

func init() {
	taskCmd.AddCommand(getCmd)
}
