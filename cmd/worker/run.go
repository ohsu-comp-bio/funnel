package worker

import (
	"context"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/rpc"
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

		// parse config file
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		// make sure server address and password is inherited by the worker
		conf = config.InheritServerProperties(conf)
		flagConf = config.InheritServerProperties(flagConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf, flagConf)
		if err != nil {
			return err
		}

		return Run(conf, taskID)
	},
}

// Run configures and runs a Worker
func Run(conf config.Config, taskID string) error {
	logger.Configure(conf.Worker.Logger)

	reader, err := rpc.NewTESClient(conf.RPC)

	if err != nil {
		return fmt.Errorf("failed to instantiate TaskReader: %v", err)
	}

	w, err := worker.NewDefaultWorker(conf.Worker)
	if err != nil {
		return err
	}

	ctx := context.Background()
	task, err := reader.FullTask(taskID)
	if err != nil {
		return err
	}

	w.Run(ctx, task)
	return nil
}
