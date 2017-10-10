package server

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/util"
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
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"github.com/ohsu-comp-bio/funnel/server/dynamodb"
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

		conf, err := util.MergeConfigFileWithFlags(configFile, flagConf)
		if err != nil {
			return fmt.Errorf("error processing config: %v", err)
		}

		return Run(context.Background(), conf)
	},
}

// Run runs a default Funnel server.
// This opens a database, and starts an API server, scheduler and task logger.
// This blocks indefinitely.
func Run(ctx context.Context, conf config.Config) error {
	logger.Configure(conf.Server.Logger)

	var backend compute.Backend
	var db server.Database
	var sched *scheduler.Scheduler
	var err error

	switch strings.ToLower(conf.Server.Database) {
	case "boltdb":
		db, err = boltdb.NewBoltDB(conf)
	case "dynamodb":
		db, err = dynamodb.NewDynamoDB(conf.Server.Databases.DynamoDB)
	}
	if err != nil {
		return fmt.Errorf("error occurred while connecting to or creating the database: %v", err)
	}

	srv := server.DefaultServer(db, conf.Server)

	switch strings.ToLower(conf.Backend) {
	case "gce", "manual", "openstack":
		sdb, ok := db.(scheduler.Database)
		if !ok {
			return fmt.Errorf("database doesn't satisfy the scheduler interface")
		}

		backend = scheduler.NewComputeBackend(sdb)

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

		sched = scheduler.NewScheduler(sdb, sbackend, conf.Scheduler)
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
