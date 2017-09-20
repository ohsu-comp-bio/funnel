package node

import (
	"context"
	"fmt"
	cmdutil "github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/spf13/cobra"
	"syscall"
)

var configFile string
var flagConf = config.Config{}

// runCmd represents the node run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Funnel node.",
	RunE: func(cmd *cobra.Command, args []string) error {

		conf, err := cmdutil.MergeConfigFileWithFlags(configFile, flagConf)
		if err != nil {
			return fmt.Errorf("error processing config: %v", err)
		}

		return Run(conf)
	},
}

func init() {
	flags := runCmd.Flags()
	flags.StringVar(&flagConf.Scheduler.Node.ID, "id", flagConf.Scheduler.Node.ID, "Node ID")
	flags.StringVar(&flagConf.Scheduler.Node.ServerAddress, "server-address", flagConf.Scheduler.Node.ServerAddress, "Address of scheduler gRPC endpoint")
	flags.DurationVar(&flagConf.Scheduler.Node.Timeout, "timeout", flagConf.Scheduler.Node.Timeout, "Timeout in seconds")
	flags.StringVar(&flagConf.Scheduler.Node.WorkDir, "work-dir", flagConf.Scheduler.Node.WorkDir, "Working Directory")
	flags.StringVar(&flagConf.Scheduler.Node.Logger.Level, "log-level", flagConf.Scheduler.Node.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Scheduler.Node.Logger.OutputFile, "log-path", flagConf.Scheduler.Node.Logger.OutputFile, "File path to write logs to")
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
}

// Run runs a node with the given config, blocking until the node exits.
func Run(conf config.Config) error {
	logger.Configure(conf.Scheduler.Node.Logger)

	if conf.Scheduler.Node.ID == "" {
		conf.Scheduler.Node.ID = scheduler.GenNodeID("manual")
	}

	n, err := scheduler.NewNode(conf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx = util.SignalContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	n.Run(ctx)

	return nil
}
