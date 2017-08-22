package worker

import (
	"context"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Funnel worker.",
	RunE: func(cmd *cobra.Command, args []string) error {

		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		workerDconf := config.WorkerInheritConfigVals(flagConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf.Worker, workerDconf)
		if err != nil {
			return err
		}

		return Start(conf)
	},
}

// Start runs a worker process with the given config, blocking until the worker exits.
func Start(conf config.Config) error {
	logger.Configure(conf.Worker.Logger)

	if conf.Worker.ID == "" {
		conf.Worker.ID = scheduler.GenWorkerID("funnel")
	}

	w, err := worker.NewWorker(conf.Worker)
	if err != nil {
		return err
	}
	w.Run(context.Background())
	return nil
}
