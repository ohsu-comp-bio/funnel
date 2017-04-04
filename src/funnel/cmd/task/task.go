package task

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

var tesServer string

// TaskCmd represents the task command
var TaskCmd = &cobra.Command{
	Use: "task",
}

func init() {
	TaskCmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
}

// responseChecker does some basic error handling and reads the response body into a byte array
func responseChecker(resp *http.Response, err error) []byte {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if (resp.StatusCode / 100) != 2 {
		fmt.Fprintf(os.Stderr, "[STATUS CODE - %d]\t%s", resp.StatusCode, body)
		os.Exit(1)
	}
	return body
}
