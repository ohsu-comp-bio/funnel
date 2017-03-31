package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "cancel a task",
	Run: func(cmd *cobra.Command, args []string) {

		for _, jobID := range args {
			u, err := url.Parse(tesServer + "/v1/jobs/" + jobID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			cli := &http.Client{}
			resp, err := cli.Do(&http.Request{
				Method: "DELETE",
				URL:    u,
			})
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
	taskCmd.AddCommand(cancelCmd)
}
