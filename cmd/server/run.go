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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	var srverr error
	go func() {
		srverr = srv.Serve(ctx)
		cancel()
	}()

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

		sched := scheduler.NewScheduler(sdb, sbackend, conf.Scheduler)
		err := sched.Start(ctx)
		if err != nil {
			return fmt.Errorf("error occurred while running the scheduler: %v", err)
		}
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
	<-ctx.Done()
	if srverr != nil {
		log.Error("Server error", srverr)
	}
	return srverr
}
