package server

import (
	"context"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/cmd/version"
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
	Short: "Runs a Funnel server.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		// parse config file
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		// make sure server address and password is inherited by scheduler nodes and workers
		conf = config.InheritServerProperties(conf)
		flagConf = config.InheritServerProperties(flagConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf, flagConf)
		if err != nil {
			return err
		}

		return Run(conf)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
	flags.StringVar(&flagConf.Server.HostName, "hostname", flagConf.Server.HostName, "Host name or IP")
	flags.StringVar(&flagConf.Server.RPCPort, "rpc-port", flagConf.Server.RPCPort, "RPC Port")
	flags.StringVar(&flagConf.Server.HTTPPort, "http-port", flagConf.Server.HTTPPort, "HTTP Port")
	flags.StringVar(&flagConf.Server.Logger.Level, "log-level", flagConf.Server.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Server.Logger.OutputFile, "log-path", flagConf.Server.Logger.OutputFile, "File path to write logs to")
	flags.StringVar(&flagConf.Server.DBPath, "db-path", flagConf.Server.DBPath, "Database path")
	flags.StringVar(&flagConf.Backend, "backend", flagConf.Backend, "Name of scheduler backend to enable")
}

// Run runs a default Funnel server.
// This opens a database, and runs an API server, scheduler and task logger.
// This blocks indefinitely.
func Run(conf config.Config) error {
	logger.Configure(conf.Server.Logger)
	version.Log(log)

	db, err := server.NewTaskBolt(conf)
	if err != nil {
		log.Error("Couldn't open database", err)
		return err
	}

	srv := server.DefaultServer(db, conf.Server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	var srverr error
	go func() {
		srverr = srv.Serve(ctx)
		cancel()
	}()

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

	backend, lerr := loader.Load(conf.Backend, conf)
	if lerr != nil {
		return lerr
	}

	sched := scheduler.NewScheduler(db, backend, conf.Scheduler)

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
