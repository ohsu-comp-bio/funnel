package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/spf13/cobra"
)

var taskID string

func init() {
	f := runCmd.Flags()
	f.StringVar(&taskID, "task-id", "", "Task ID")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a task directly, bypassing the server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if taskID == "" {
			fmt.Printf("No taskID was provided.\n\n")
			return cmd.Help()
		}

		conf, err := util.MergeConfigFileWithFlags(configFile, flagConf)
		if err != nil {
			return fmt.Errorf("error processing config: %v", err)
		}

		return Run(conf.Worker, taskID)
	},
}

// Run configures and runs a Worker
func Run(conf config.Worker, taskID string) error {
	logger.Configure(conf.Logger)
	w, err := worker.NewDefaultWorker(conf, taskID)
	if err != nil {
		return err
	}
	w.Run(context.Background())
	return nil
}
