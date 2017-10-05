package gce

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/node"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/compute/gce"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

var log = logger.New("gce cmd")

// Cmd represents the 'funnel gce" CLI command set.
var Cmd = &cobra.Command{
	Use: "gce",
}

func init() {
	Cmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := config.DefaultConfig()

		// Check that this is a GCE VM environment.
		// If not, fail.
		meta, merr := gce.LoadMetadata()
		if merr != nil {
			log.Error("Error getting GCE metadata", merr)
			return fmt.Errorf("can't find GCE metadata. This command requires a GCE environment")
		}

		log.Info("Loaded GCE metadata")
		log.Debug("GCE metadata", meta)

		var err error
		conf, err = gce.WithMetadataConfig(conf, meta)
		if err != nil {
			return err
		}

		if conf.Scheduler.Node.ID != "" {
			logger.Configure(conf.Scheduler.Node.Logger)
			return node.Run(conf)
		}

		logger.Configure(conf.Server.Logger)
		ctx := context.Background()
		return server.Run(ctx, conf)
	},
}
