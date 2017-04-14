package cmd

import (
	"funnel/config"
	"funnel/scheduler"
	"funnel/worker"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

var workerConfigFile string
var workerBaseConf = config.Config{}

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := config.DefaultConfig()

		if workerConfigFile != "" {
			config.LoadConfigOrExit(workerConfigFile, &conf)
		}

		workerDconf := config.WorkerInheritConfigVals(workerBaseConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf.Worker, workerDconf)
		if err != nil {
			return err
		}

		initLogging(conf)

		if conf.Worker.ID == "" {
			conf.Worker.ID = scheduler.GenWorkerID("funnel")
		}

		w, err := worker.NewWorker(conf.Worker)
		if err != nil {
			return err
		}
		w.Run()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
	workerCmd.Flags().StringVar(&workerBaseConf.Worker.ID, "id", workerBaseConf.Worker.ID, "Worker ID")
	workerCmd.Flags().DurationVar(&workerBaseConf.Worker.Timeout, "timeout", workerBaseConf.Worker.Timeout, "Timeout in seconds")
	workerCmd.Flags().StringVarP(&workerConfigFile, "config", "c", "", "Config File")
	workerCmd.Flags().StringVar(&workerBaseConf.HostName, "hostname", workerBaseConf.HostName, "Host name or IP")
	workerCmd.Flags().StringVar(&workerBaseConf.RPCPort, "rpc-port", workerBaseConf.RPCPort, "RPC Port")
	workerCmd.Flags().StringVar(&workerBaseConf.WorkDir, "work-dir", workerBaseConf.WorkDir, "Working Directory")
	workerCmd.Flags().StringVar(&workerBaseConf.LogLevel, "log-level", workerBaseConf.LogLevel, "Level of logging")
	workerCmd.Flags().StringVar(&workerBaseConf.LogPath, "log-path", workerBaseConf.LogLevel, "File path to write logs to")
}
