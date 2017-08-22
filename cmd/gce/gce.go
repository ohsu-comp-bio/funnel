package gce

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/scheduler/gce"
	"github.com/spf13/cobra"
)

var log = logger.New("gce cmd")

// Cmd represents the 'funnel gce" CLI command set.
var Cmd = &cobra.Command{
	Use: "gce",
}

func init() {
	Cmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use: "start",
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

		if conf.Worker.ID != "" {
			logger.Configure(conf.Worker.Logger)
			return worker.Start(conf)
		}

		logger.Configure(conf.Worker.Logger)
		return server.Start(conf)
	},
}
