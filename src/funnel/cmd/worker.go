package cmd

import (
	"funnel/config"
	"funnel/logger"
	"funnel/scheduler"
	"funnel/worker"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"os"
)

var workerLog = logger.New("funnel-worker")
var workerDconf = baseConf.Worker

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var wconf = config.WorkerDefaultConfig(config.DefaultConfig())
		var conf config.Config
		config.LoadConfigOrExit(configFile, &conf)
		wconf = conf.Worker

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&wconf, workerDconf)
		if err != nil {
			panic(err)
		}

		startWorker(wconf)
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
	workerCmd.Flags().StringVar(&workerDconf.ID, "id", workerDconf.ID, "Worker ID")
	workerCmd.Flags().DurationVar(&workerDconf.Timeout, "timeout", workerDconf.Timeout, "Timeout in seconds")
}

func startWorker(conf config.Worker) {
	logger.SetLevel(conf.LogLevel)

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems?
	// Configure logging
	if conf.LogPath != "" {
		logFile, err := os.OpenFile(conf.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			workerLog.Error("Can't open log output file", "path", conf.LogPath)
		} else {
			logger.SetOutput(logFile)
		}
	}

	if conf.ID == "" {
		conf.ID = scheduler.GenWorkerID()
	}

	w, err := worker.NewWorker(conf)
	if err != nil {
		workerLog.Error("Can't create worker", err)
		return
	}
	w.Run()
}
