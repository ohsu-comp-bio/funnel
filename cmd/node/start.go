package node

import (
	"context"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/spf13/cobra"
)

var configFile string
var flagConf = config.Config{}

// startCmd represents the node start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Funnel node.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// parse config file
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		// make sure server address and password is inherited by scheduler nodes and workers
		conf.InheritServerProperties()
		flagConf.InheritServerProperties()

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf, flagConf)
		if err != nil {
			return err
		}

		return Start(conf)
	},
}

func init() {
	flags := startCmd.Flags()
	flags.StringVar(&flagConf.Scheduler.Node.ID, "id", flagConf.Scheduler.Node.ID, "Node ID")
	flags.StringVar(&flagConf.Scheduler.Node.ServerAddress, "server-address", flagConf.Scheduler.Node.ServerAddress, "Address of scheduler gRPC endpoint")
	flags.DurationVar(&flagConf.Scheduler.Node.Timeout, "timeout", flagConf.Scheduler.Node.Timeout, "Timeout in seconds")
	flags.StringVar(&flagConf.Scheduler.Node.WorkDir, "work-dir", flagConf.Scheduler.Node.WorkDir, "Working Directory")
	flags.StringVar(&flagConf.Scheduler.Node.Logger.Level, "log-level", flagConf.Scheduler.Node.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Scheduler.Node.Logger.OutputFile, "log-path", flagConf.Scheduler.Node.Logger.OutputFile, "File path to write logs to")
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
}

// Start starts a node with the given config, blocking until the node exits.
func Start(conf config.Config) error {
	logger.Configure(conf.Scheduler.Node.Logger)

	if conf.Scheduler.Node.ID == "" {
		conf.Scheduler.Node.ID = scheduler.GenNodeID("funnel")
	}

	n, err := scheduler.NewNode(conf)
	if err != nil {
		return err
	}
	n.Start(context.Background())
	return nil
}
