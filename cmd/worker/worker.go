package worker

import (
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/spf13/cobra"
)

var configFile string
var flagConf = config.Config{}

// Cmd represents the worker command
var Cmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"workers"},
	Short:   "Starts a Funnel worker.",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		workerDconf := config.WorkerInheritConfigVals(flagConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf.Worker, workerDconf)
		if err != nil {
			return err
		}

		return Run(conf)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&flagConf.Worker.ID, "id", flagConf.Worker.ID, "Worker ID")
	flags.DurationVar(&flagConf.Worker.Timeout, "timeout", flagConf.Worker.Timeout, "Timeout in seconds")
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
	flags.StringVar(&flagConf.HostName, "hostname", flagConf.HostName, "Host name or IP")
	flags.StringVar(&flagConf.RPCPort, "rpc-port", flagConf.RPCPort, "RPC Port")
	flags.StringVar(&flagConf.WorkDir, "work-dir", flagConf.WorkDir, "Working Directory")
	flags.StringVar(&flagConf.Logger.Level, "log-level", flagConf.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Logger.OutputFile, "log-path", flagConf.Logger.OutputFile, "File path to write logs to")
}

// Run runs a worker with the given config, blocking until the worker exits.
func Run(conf config.Config) error {

	logger.Configure(conf.Logger)

	if conf.Worker.ID == "" {
		conf.Worker.ID = scheduler.GenWorkerID("funnel")
	}

	w, err := worker.NewWorker(conf.Worker)
	if err != nil {
		return err
	}
	w.Run()
	return nil
}
