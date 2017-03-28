package cmd

import (
	"funnel/config"
	"funnel/logger"
	"funnel/scheduler"
	"funnel/scheduler/condor"
	"funnel/scheduler/gce"
	"funnel/scheduler/local"
	"funnel/scheduler/openstack"
	"funnel/server"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var serverLog = logger.New("funnel-server")

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var conf = config.DefaultConfig()
		if configFile != "" {
			config.LoadConfigOrExit(configFile, &conf)
		}

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf, baseConf)
		if err != nil {
			panic(err)
		}

		// make sure the proper defaults are set
		conf.Worker = config.WorkerInheritConfigVals(conf)

		startServer(conf)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&baseConf.HTTPPort, "http-port", baseConf.HTTPPort, "HTTP Port")
	serverCmd.Flags().StringVar(&baseConf.DBPath, "db-path", baseConf.DBPath, "Database path")
	serverCmd.Flags().StringVar(&baseConf.Scheduler, "scheduler", baseConf.Scheduler, "Name of scheduler to enable")
}

func startServer(conf config.Config) {
	var err error

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

	err = os.MkdirAll(conf.WorkDir, 0755)
	if err != nil {
		panic(err)
	}

	taski, err := server.NewTaskBolt(conf)
	if err != nil {
		serverLog.Error("Couldn't open database", err)
		return
	}

	s := server.NewGA4GHServer()
	s.RegisterTaskServer(taski)
	s.RegisterScheduleServer(taski)
	go s.Start(conf.RPCPort)

	var sched scheduler.Scheduler
	switch strings.ToLower(conf.Scheduler) {
	case "local":
		// TODO worker will stay alive if the parent process panics
		sched, err = local.NewScheduler(conf)
	case "condor":
		sched = condor.NewScheduler(conf)
	case "gce":
		sched, err = gce.NewScheduler(conf)
	case "openstack":
		sched, err = openstack.NewScheduler(conf)
	default:
		serverLog.Error("Unknown scheduler", "scheduler", conf.Scheduler)
		return
	}

	if err != nil {
		serverLog.Error("Couldn't create scheduler", err)
		return
	}

	go scheduler.ScheduleLoop(taski, sched, conf)

	// If the scheduler implements the Scaler interface,
	// start a scaler loop
	if s, ok := sched.(scheduler.Scaler); ok {
		go scheduler.ScaleLoop(taski, s, conf)
	}

	// TODO if port 8000 is already busy, does this lock up silently?
	server.StartHTTPProxy(conf.HostName+":"+conf.RPCPort, conf.HTTPPort, conf.ContentDir)
}
