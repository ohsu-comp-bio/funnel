package cmd

import (
	"context"
	"funnel/config"
	"funnel/logger"
	"funnel/scheduler"
	// Register scheduler backends
	_ "funnel/scheduler/condor"
	_ "funnel/scheduler/gce"
	_ "funnel/scheduler/local"
	_ "funnel/scheduler/openstack"
	"funnel/server"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"os"
)

var configFile string
var baseConf = config.Config{}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a Funnel server.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var conf = config.DefaultConfig()
		if configFile != "" {
			config.LoadConfigOrExit(configFile, &conf)
		}

		// file vals <- cli val
		err = mergo.MergeWithOverwrite(&conf, baseConf)
		if err != nil {
			return err
		}

		// make sure the proper defaults are set
		conf.Worker = config.WorkerInheritConfigVals(conf)

		initLogging(conf)

		db, err := server.NewTaskBolt(conf)
		if err != nil {
			logger.Error("Couldn't open database", err)
			return err
		}

		srv, err := server.NewServer(db, conf)
		if err != nil {
			return err
		}

		sched, err := scheduler.NewScheduler(db, conf)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start server
		srv.Start(ctx)

		// Start scheduler
		err = sched.Start(ctx)
		if err != nil {
			return err
		}

		// Block
		<-ctx.Done()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config File")
	serverCmd.Flags().StringVar(&baseConf.HostName, "hostname", baseConf.HostName, "Host name or IP")
	serverCmd.Flags().StringVar(&baseConf.RPCPort, "rpc-port", baseConf.RPCPort, "RPC Port")
	serverCmd.Flags().StringVar(&baseConf.WorkDir, "work-dir", baseConf.WorkDir, "Working Directory")
	serverCmd.Flags().StringVar(&baseConf.LogLevel, "log-level", baseConf.LogLevel, "Level of logging")
	serverCmd.Flags().StringVar(&baseConf.LogPath, "log-path", baseConf.LogLevel, "File path to write logs to")
	serverCmd.Flags().StringVar(&baseConf.HTTPPort, "http-port", baseConf.HTTPPort, "HTTP Port")
	serverCmd.Flags().StringVar(&baseConf.DBPath, "db-path", baseConf.DBPath, "Database path")
	serverCmd.Flags().StringVar(&baseConf.Scheduler, "scheduler", baseConf.Scheduler, "Name of scheduler to enable")
}

func initLogging(conf config.Config) {
	logger.SetLevel(conf.LogLevel)

	// TODO Good defaults, configuration, and reusable way to configure logging.
	//      Also, how do we get this to default to /var/log/tes/worker.log
	//      without having file permission problems? syslog?
	// Configure logging
	if conf.LogPath != "" {
		logFile, err := os.OpenFile(
			conf.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666,
		)
		if err != nil {
			logger.Error("Can't open log output file", "path", conf.LogPath)
		} else {
			logger.SetOutput(logFile)
		}
	}
}
