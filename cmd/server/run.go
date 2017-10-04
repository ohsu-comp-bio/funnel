package server

import (
	"context"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/compute/gce"
	"github.com/ohsu-comp-bio/funnel/compute/gridengine"
	"github.com/ohsu-comp-bio/funnel/compute/htcondor"
	"github.com/ohsu-comp-bio/funnel/compute/local"
	"github.com/ohsu-comp-bio/funnel/compute/manual"
	"github.com/ohsu-comp-bio/funnel/compute/openstack"
	"github.com/ohsu-comp-bio/funnel/compute/pbs"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/compute/slurm"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/elastic"
	"github.com/spf13/cobra"
	"strings"
)

var log = logger.New("server run cmd")

// runCmd represents the `funnel server run` command.
var runCmd = &cobra.Command{
	Use:   "run",
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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return Run(ctx, conf)
	},
}

// Run runs a default Funnel server.
// This opens a database, and starts an API server, scheduler and task logger.
func Run(ctx context.Context, conf config.Config) error {
	logger.Configure(conf.Server.Logger)

	var backend compute.Backend
	var sched *scheduler.Scheduler

  db, eserr := elastic.NewTES(elastic.DefaultConfig())
  db.Init(ctx)
	if eserr != nil {
		return fmt.Errorf("error occurred while connecting to or creating the database: %v", eserr)
	}

	srv := server.DefaultServer(db, conf.Server)

	var err error
	switch strings.ToLower(conf.Backend) {
	case "gce", "manual", "openstack":

		backend = scheduler.NewComputeBackend(db)

		var sbackend scheduler.Backend
		switch strings.ToLower(conf.Backend) {
		case "gce":
			sbackend, err = gce.NewBackend(conf)
		case "manual":
			sbackend, err = manual.NewBackend(conf)
		case "openstack":
			sbackend, err = openstack.NewBackend(conf)
		}
		if err != nil {
			return fmt.Errorf("error occurred while setting up backend: %v", err)
		}

		sched = scheduler.NewScheduler(db, sbackend, conf.Scheduler)
	case "gridengine":
		backend = gridengine.NewBackend(conf)
	case "htcondor":
		backend = htcondor.NewBackend(conf)
	case "local":
		backend = local.NewBackend(conf)
	case "pbs":
		backend = pbs.NewBackend(conf)
	case "slurm":
		backend = slurm.NewBackend(conf)
	default:
		return fmt.Errorf("unknown backend")
	}

	db.WithComputeBackend(backend)

	// Block

	// Start server
	errch := make(chan error)
	go func() {
		errch <- srv.Serve(ctx)
	}()

	// Start Scheduler
	if sched != nil {
		go func() {
			errch <- sched.Run(ctx)
		}()
	}

	// Block until done.
	// Server and scheduler must be stopped via the context.
	return <-errch
}
