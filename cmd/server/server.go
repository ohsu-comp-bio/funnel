package server

import (
	"context"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/scheduler/gce"
	"github.com/ohsu-comp-bio/funnel/scheduler/gridengine"
	"github.com/ohsu-comp-bio/funnel/scheduler/htcondor"
	"github.com/ohsu-comp-bio/funnel/scheduler/local"
	"github.com/ohsu-comp-bio/funnel/scheduler/manual"
	"github.com/ohsu-comp-bio/funnel/scheduler/openstack"
	"github.com/ohsu-comp-bio/funnel/scheduler/pbs"
	"github.com/ohsu-comp-bio/funnel/scheduler/slurm"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/spf13/cobra"
)

var log = logger.New("server cmd")
var configFile string
var flagConf = config.Config{}

// Cmd represents the `funnel server` CLI command set.
var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a Funnel server.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var conf = config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		// file vals <- cli val
		err = mergo.MergeWithOverwrite(&conf, flagConf)
		if err != nil {
			return err
		}

		return Run(conf)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
	flags.StringVar(&flagConf.HostName, "hostname", flagConf.HostName, "Host name or IP")
	flags.StringVar(&flagConf.RPCPort, "rpc-port", flagConf.RPCPort, "RPC Port")
	flags.StringVar(&flagConf.WorkDir, "work-dir", flagConf.WorkDir, "Working Directory")
	flags.StringVar(&flagConf.Logger.Level, "log-level", flagConf.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Logger.OutputFile, "log-path", flagConf.Logger.OutputFile, "File path to write logs to")
	flags.StringVar(&flagConf.HTTPPort, "http-port", flagConf.HTTPPort, "HTTP Port")
	flags.StringVar(&flagConf.DBPath, "db-path", flagConf.DBPath, "Database path")
	flags.StringVar(&flagConf.Scheduler, "scheduler", flagConf.Scheduler, "Name of scheduler to enable")
}

// Run runs a default Funnel server.
// This opens a database, and starts an API server and scheduler.
// This blocks indefinitely.
func Run(conf config.Config) error {
	logger.Configure(conf.Logger)

	// make sure the proper defaults are set
	conf.Worker = config.WorkerInheritConfigVals(conf)

	db, err := server.NewTaskBolt(conf)
	if err != nil {
		log.Error("Couldn't open database", err)
		return err
	}

	loader := scheduler.BackendLoader{
		gce.Name:        gce.NewBackend,
		htcondor.Name:   htcondor.NewBackend,
		openstack.Name:  openstack.NewBackend,
		local.Name:      local.NewBackend,
		manual.Name:     manual.NewBackend,
		pbs.Name:        pbs.NewBackend,
		gridengine.Name: gridengine.NewBackend,
		slurm.Name:      slurm.NewBackend,
	}

	backend, lerr := loader.Load(conf.Scheduler, conf)
	if lerr != nil {
		return lerr
	}

	srv := server.DefaultServer(db, conf)
	sched := scheduler.NewScheduler(db, backend, conf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	var srverr error
	go func() {
		srverr = srv.Serve(ctx)
		cancel()
	}()

	// Start scheduler
	err = sched.Start(ctx)
	if err != nil {
		return err
	}

	// Block
	<-ctx.Done()
	if srverr != nil {
		log.Error("Server error", srverr)
	}
	return srverr
}
