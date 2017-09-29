package batch

import (
	"fmt"
	"github.com/spf13/cobra"
)

var defaultsCmd = &cobra.Command{
	Use:     "default",
	Aliases: []string{"defaults"},
	Short:   "Print the default compute environment and job queue",
	RunE: func(cmd *cobra.Command, args []string) error {

		cli, err := newBatchSvc(conf, true)
		if err != nil {
			return err
		}

		fmt.Println("ComputEnvironment:")
		_, err = cli.CreateComputeEnvironment()
		if err != nil {
			return err
		}
		fmt.Println("")

		fmt.Println("JobQueue:")
		_, err = cli.CreateJobQueue()
		if err != nil {
			return err
		}
		fmt.Println("")

		return nil
	},
}
